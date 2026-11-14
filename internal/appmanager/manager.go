// Package appmanager supervises private local app processes. Loading/validating
// a Project never calls it; only actual dev/host startup may launch an app.
package appmanager

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"flea/internal/appapi"
	"flea/internal/appbundle"
	"flea/internal/appruntime"
	"flea/internal/sitefiles"
)

// A single verified executable inode is shared by every site in this process.
var bundles struct {
	sync.Mutex
	dir, path string
	refs      int
}

// Spec describes one managed local app. Bundle selects the embedded Contact
// executable; otherwise Executable is an already-confined absolute path whose
// validated content identity (Hash, Size) is re-checked when it is frozen for
// launch. RuntimeRoot is the trusted per-site runtime anchor; Storage is the
// app-owned private directory beneath it that is handed to the child.
type Spec struct {
	Name        string
	Executable  string
	Bundle      bool
	RuntimeRoot string
	Storage     string
	Capability  http.Handler
	Hash        [32]byte
	Size        int64
	Version     string // manifest version; empty skips the handshake version check
}

func acquire(ctx context.Context) (string, func(), error) {
	bundles.Lock()
	defer bundles.Unlock()
	if bundles.refs == 0 {
		dir, err := os.MkdirTemp("", "flea-contact-")
		if err != nil {
			return "", nil, err
		}
		name := filepath.Join(dir, "contact")
		if err := appbundle.Extract(ctx, name); err != nil {
			os.Remove(dir)
			return "", nil, err
		}
		bundles.dir, bundles.path = dir, name
	}
	bundles.refs++
	return bundles.path, func() {
		bundles.Lock()
		defer bundles.Unlock()
		bundles.refs--
		if bundles.refs == 0 {
			os.Remove(bundles.path)
			os.Remove(bundles.dir)
			bundles.path, bundles.dir = "", ""
		}
	}, nil
}

// executableFor returns the confined executable for a spec and a release func.
// A custom executable is re-opened through a confined root, verified against the
// identity recorded at load time, and copied to a private 0500 file; only the
// frozen copy is ever launched, so a parent-directory or file swap after
// validation cannot change what runs.
func executableFor(ctx context.Context, spec Spec) (string, func(), error) {
	if spec.Bundle {
		return acquire(ctx)
	}
	dir, err := os.MkdirTemp("", "flea-app-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(dir) }
	frozen, err := freezeExecutable(spec, dir)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	return frozen, cleanup, nil
}

func freezeExecutable(spec Spec, dir string) (string, error) {
	root, err := os.OpenRoot(filepath.Dir(spec.Executable))
	if err != nil {
		return "", fmt.Errorf("%s executable is unavailable", spec.Name)
	}
	defer root.Close()
	base := filepath.Base(spec.Executable)
	info, err := root.Lstat(base)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("%s executable is not a runnable regular file", spec.Name)
	}
	source, err := root.Open(base)
	if err != nil {
		return "", fmt.Errorf("%s executable is unavailable", spec.Name)
	}
	defer source.Close()
	actual, err := source.Stat()
	if err != nil || !os.SameFile(info, actual) || !actual.Mode().IsRegular() || actual.Size() != spec.Size {
		return "", fmt.Errorf("%s executable changed after validation", spec.Name)
	}
	// A hardlinked executable could be rewritten through its other name after
	// the identity check, so it is rejected on the frozen copy as well as at
	// load. The confined root above prevents a parent-directory swap.
	if sitefiles.Hardlinked(info) || sitefiles.Hardlinked(actual) {
		return "", fmt.Errorf("%s executable must have a single link", spec.Name)
	}
	frozen := filepath.Join(dir, "app")
	target, err := os.OpenFile(frozen, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o500)
	if err != nil {
		return "", fmt.Errorf("%s executable could not be prepared", spec.Name)
	}
	h := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(target, h), io.LimitReader(source, spec.Size+1))
	syncErr := target.Sync()
	closeErr := target.Close()
	if copyErr != nil || syncErr != nil || closeErr != nil || n != spec.Size {
		return "", fmt.Errorf("%s executable could not be prepared", spec.Name)
	}
	if !bytes.Equal(h.Sum(nil), spec.Hash[:]) {
		return "", fmt.Errorf("%s executable changed after validation", spec.Name)
	}
	return frozen, nil
}

// prepareStorage creates the app's private storage at the only supported
// layout, <runtime>/apps/<name>, pinned to a trusted anchor so a symlinked
// ancestor cannot redirect state outside the site. Existing runtime and app
// directories must be private and service-owned; a directory this call did not
// create is never chmodded, it is rejected.
func prepareStorage(spec Spec) error {
	if spec.Storage == "" {
		return nil
	}
	if !filepath.IsAbs(spec.Storage) || !filepath.IsAbs(spec.RuntimeRoot) {
		return fmt.Errorf("%s storage path must be absolute", spec.Name)
	}
	anchor := filepath.Dir(filepath.Dir(spec.RuntimeRoot))
	if filepath.Clean(spec.RuntimeRoot) != filepath.Join(anchor, "flea", "runtime") {
		return fmt.Errorf("%s storage path is not the managed runtime", spec.Name)
	}
	if filepath.Clean(spec.Storage) != filepath.Join(spec.RuntimeRoot, "apps", spec.Name) {
		return fmt.Errorf("%s storage path is outside the app runtime", spec.Name)
	}
	info, err := os.Lstat(anchor)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s storage anchor is unavailable", spec.Name)
	}
	root, err := os.OpenRoot(anchor)
	if err != nil {
		return fmt.Errorf("%s storage is unavailable", spec.Name)
	}
	defer root.Close()
	// flea is the administrator-owned container; runtime, apps and the app leaf
	// must be private and owned by this service user. Each component is checked
	// through the confined root, which never follows a link outside the anchor.
	parts := []struct {
		name    string
		private bool
	}{
		{"flea", false},
		{"runtime", true},
		{"apps", true},
		{spec.Name, true},
	}
	relative := ""
	for _, part := range parts {
		relative = filepath.Join(relative, part.name)
		existing, statErr := root.Lstat(relative)
		switch {
		case statErr == nil:
			if !existing.IsDir() || existing.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("%s storage is not a real directory", spec.Name)
			}
			if part.private {
				if existing.Mode().Perm()&0o077 != 0 || !ownedByService(existing) {
					return fmt.Errorf("%s storage is not a private service directory", spec.Name)
				}
			} else if existing.Mode().Perm()&0o022 != 0 {
				return fmt.Errorf("%s storage is writable by other users", spec.Name)
			}
		case os.IsNotExist(statErr):
			if err := root.Mkdir(relative, 0o700); err != nil {
				return fmt.Errorf("%s storage directory is unavailable", spec.Name)
			}
			// Set the mode explicitly so a restrictive umask cannot leave the
			// service user without access to a directory this call created.
			if err := root.Chmod(relative, 0o700); err != nil {
				return fmt.Errorf("%s storage directory is unavailable", spec.Name)
			}
		default:
			return fmt.Errorf("%s storage is unavailable", spec.Name)
		}
	}
	return nil
}

type Manager struct {
	mu     sync.RWMutex
	client *appapi.Client
	cancel context.CancelFunc
	done   chan struct{}
}

func token() (string, error) {
	var b [32]byte
	_, err := rand.Read(b[:])
	return hex.EncodeToString(b[:]), err
}

func Start(ctx context.Context, spec Spec, logger *log.Logger) (*Manager, error) {
	if !appapi.ValidName(spec.Name) {
		return nil, fmt.Errorf("invalid app name")
	}
	if err := prepareStorage(spec); err != nil {
		return nil, err
	}
	executable, release, err := executableFor(ctx, spec)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	m := &Manager{cancel: cancel, done: make(chan struct{})}
	go func() { defer close(m.done); defer release(); m.run(ctx, spec, executable, logger) }()
	return m, nil
}

func (m *Manager) Client() *appapi.Client { m.mu.RLock(); defer m.mu.RUnlock(); return m.client }
func (m *Manager) Close()                 { m.cancel(); <-m.done }

func (m *Manager) logf(logger *log.Logger, name, format string, args ...any) {
	if logger != nil {
		logger.Printf("%s "+format, append([]any{name}, args...)...)
	}
}

func (m *Manager) run(ctx context.Context, spec Spec, executable string, logger *log.Logger) {
	for attempt := 0; attempt < 5 && ctx.Err() == nil; attempt++ {
		if attempt > 0 {
			m.logf(logger, spec.Name, "stopped. Restarting its private service.")
			timer := time.NewTimer(time.Duration(1<<(attempt-1)) * time.Second)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
		_ = m.once(ctx, spec, executable, logger)
	}
	if ctx.Err() == nil && logger != nil {
		m.logf(logger, spec.Name, "could not stay running. Other pages remain available; restart ItsyWeb after checking the app.")
	}
}

func (m *Manager) once(ctx context.Context, spec Spec, executable string, logger *log.Logger) error {
	credential, err := token()
	if err != nil {
		return err
	}
	capToken, err := token()
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	inherited, err := listener.(*net.TCPListener).File()
	if err != nil {
		return err
	}
	defer inherited.Close()
	var callback net.Listener
	var capabilityURL string
	if spec.Capability != nil {
		callback, err = net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			return err
		}
		defer callback.Close()
		capabilityURL = "http://" + callback.Addr().String()
		server := &http.Server{ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 16 << 10, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != appruntime.CapabilityPath || r.URL.RawPath != "" || r.URL.RawQuery != "" {
				http.NotFound(w, r)
				return
			}
			if len(r.Header.Values("Authorization")) != 1 || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+capToken)) != 1 {
				http.Error(w, "Unauthorized", 401)
				return
			}
			spec.Capability.ServeHTTP(w, r)
		})}
		go server.Serve(callback)
		defer server.Close()
	}
	configRead, configWrite, err := os.Pipe()
	if err != nil {
		return err
	}
	defer configRead.Close()
	defer configWrite.Close()
	aliveRead, aliveWrite, err := os.Pipe()
	if err != nil {
		return err
	}
	defer aliveRead.Close()
	defer aliveWrite.Close()
	cmd := exec.Command(executable)
	cmd.Env = []string{"LANG=C"}
	cmd.Dir = filepath.Dir(executable)
	cmd.ExtraFiles = []*os.File{inherited, configRead, aliveRead}
	// Child output is discarded, never forwarding potentially private payloads.
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s could not start", spec.Name)
	}
	configRead.Close()
	aliveRead.Close()
	inherited.Close()
	listener.Close()
	wait := make(chan error, 1)
	go func() { wait <- cmd.Wait() }()
	stopped := false
	defer func() {
		m.mu.Lock()
		client := m.client
		m.client = nil
		m.mu.Unlock()
		if client != nil {
			client.Close()
		}
		aliveWrite.Close()
		if !stopped {
			select {
			case <-wait:
			case <-time.After(3 * time.Second):
				cmd.Process.Kill()
				<-wait
			}
		}
	}()
	if err := json.NewEncoder(configWrite).Encode(struct {
		Name, Token, Storage, CapabilityURL, CapabilityToken string
	}{spec.Name, credential, spec.Storage, capabilityURL, capToken}); err != nil {
		return err
	}
	configWrite.Close()
	client, err := appapi.New("http://"+listener.Addr().String(), spec.Name, credential)
	if err != nil {
		return err
	}
	defer client.Close()
	ready, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	for {
		if identity, handshakeErr := client.HandshakeIdentity(ready); handshakeErr == nil {
			// The running service must match the package manifest it was
			// launched from, not merely the name and protocol.
			if spec.Version != "" && identity.Version != spec.Version {
				return fmt.Errorf("%s version mismatch", spec.Name)
			}
			break
		}
		select {
		case <-ready.Done():
			return fmt.Errorf("%s handshake failed", spec.Name)
		case <-time.After(50 * time.Millisecond):
		case err := <-wait:
			stopped = true
			return err
		}
	}
	m.mu.Lock()
	m.client = client
	m.mu.Unlock()
	select {
	case <-ctx.Done():
		return nil
	case err := <-wait:
		stopped = true
		return err
	}
}
