// Package appmanager supervises private local app processes. Loading/validating
// a Project never calls it; only actual dev/host startup may launch an app.
package appmanager

import (
	"context"
	"crypto/rand"
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
)

// A single verified executable inode is shared by every site in this process.
var bundles struct {
	sync.Mutex
	dir, path string
	refs      int
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

func Start(ctx context.Context, capability http.Handler, logger *log.Logger) (*Manager, error) {
	executable, release, err := acquire(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	m := &Manager{cancel: cancel, done: make(chan struct{})}
	go func() { defer close(m.done); defer release(); m.run(ctx, executable, capability, logger) }()
	return m, nil
}

func (m *Manager) Client() *appapi.Client { m.mu.RLock(); defer m.mu.RUnlock(); return m.client }
func (m *Manager) Close()                 { m.cancel(); <-m.done }

func (m *Manager) run(ctx context.Context, executable string, capability http.Handler, logger *log.Logger) {
	for attempt := 0; attempt < 5 && ctx.Err() == nil; attempt++ {
		if attempt > 0 {
			if logger != nil {
				logger.Print("Contact stopped. Restarting its private service.")
			}
			timer := time.NewTimer(time.Duration(1<<(attempt-1)) * time.Second)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
		_ = m.once(ctx, executable, capability)
	}
	if ctx.Err() == nil && logger != nil {
		logger.Print("Contact could not stay running. Other pages remain available; restart Flea after checking the app.")
	}
}

func (m *Manager) once(ctx context.Context, executable string, capability http.Handler) error {
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
	callback, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer callback.Close()
	server := &http.Server{ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 16 << 10, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != appruntime.CapabilityPath || r.URL.RawPath != "" || r.URL.RawQuery != "" {
			http.NotFound(w, r)
			return
		}
		if len(r.Header.Values("Authorization")) != 1 || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+capToken)) != 1 {
			http.Error(w, "Unauthorized", 401)
			return
		}
		capability.ServeHTTP(w, r)
	})}
	go server.Serve(callback)
	defer server.Close()
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
		return fmt.Errorf("Contact could not start")
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
	if err := json.NewEncoder(configWrite).Encode(struct{ Token, CapabilityURL, CapabilityToken string }{credential, "http://" + callback.Addr().String(), capToken}); err != nil {
		return err
	}
	configWrite.Close()
	client, err := appapi.New("http://"+listener.Addr().String(), "contact", credential)
	if err != nil {
		return err
	}
	defer client.Close()
	ready, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	for {
		if client.Handshake(ready) == nil {
			break
		}
		select {
		case <-ready.Done():
			return fmt.Errorf("Contact handshake failed")
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
