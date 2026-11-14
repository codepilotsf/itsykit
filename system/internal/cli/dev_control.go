package cli

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"itsy/internal/site"
)

var errDevReplaced = errors.New("dev server replaced")

// The control socket is private to this OS user, never part of the HTTP server
// or site tree. Only dev offers this endpoint; pull/reset cannot be stopped.
func devControlPath(root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	// macOS limits Unix socket paths to 104 bytes. Its per-user TempDir can
	// already consume most of that; use a short, owner-checked private directory.
	dir := filepath.Join("/tmp", fmt.Sprintf("flea-dev-%d", os.Getuid()))
	if err := os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
		return "", err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || info.Mode().Perm() != 0700 || !ok || stat.Uid != uint32(os.Getuid()) {
		return "", fmt.Errorf("development control directory must be private and owned by your user: %s", dir)
	}
	// Match the existing lock's per-TempDir scope as well as its canonical root.
	hash := sha256.Sum256([]byte(filepath.Clean(os.TempDir()) + "\x00" + root))
	return filepath.Join(dir, fmt.Sprintf("%x.sock", hash)), nil
}

func beginDevelopment(ctx context.Context, root string, out io.Writer) (context.Context, func(), error) {
	address, err := devControlPath(root)
	if err != nil {
		return nil, nil, err
	}
	deadline := time.Now().Add(8 * time.Second)
	discoveryDeadline := time.Now().Add(time.Second)
	requested := false
	for {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		unlock, err := site.LockDevelopment(root)
		if err == nil {
			return serveDevControl(ctx, address, unlock)
		}
		if !errors.Is(err, site.ErrDevelopmentBusy) {
			return nil, nil, err
		}
		if !requested {
			requested = requestDevStop(ctx, address)
			if requested {
				fmt.Fprintln(out, "Stopping the previous dev server for this site...")
			} else if time.Now().After(discoveryDeadline) {
				return nil, nil, fmt.Errorf("this site is busy with pull, reset, or a dev server that cannot be restarted automatically; let that command finish or stop it in its terminal, then run itsy dev again")
			}
		}
		if time.Now().After(deadline) {
			return nil, nil, fmt.Errorf("the previous dev server has not finished stopping; wait a moment, then run itsy dev again")
		}
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func requestDevStop(ctx context.Context, address string) bool {
	conn, err := (&net.Dialer{Timeout: 200 * time.Millisecond}).DialContext(ctx, "unix", address)
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(time.Second))
	if _, err := io.WriteString(conn, "stop\n"); err != nil {
		return false
	}
	var reply [9]byte
	_, err = io.ReadFull(conn, reply[:])
	return err == nil && string(reply[:]) == "stopping\n"
}

func serveDevControl(ctx context.Context, address string, unlock func()) (context.Context, func(), error) {
	// Holding the site lock proves no other dev instance owns this endpoint.
	// Remove only a stale socket, never a regular file or a symlink.
	if info, err := os.Lstat(address); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			unlock()
			return nil, nil, fmt.Errorf("unexpected file at development control socket: %s", address)
		}
		if err := os.Remove(address); err != nil {
			unlock()
			return nil, nil, err
		}
	} else if !os.IsNotExist(err) {
		unlock()
		return nil, nil, err
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: address, Net: "unix"})
	if err != nil {
		unlock()
		return nil, nil, err
	}
	devCtx, cancel := context.WithCancelCause(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.AcceptUnix()
			if err != nil {
				return
			}
			_ = conn.SetDeadline(time.Now().Add(time.Second))
			var request [5]byte
			if _, err := io.ReadFull(conn, request[:]); err == nil && string(request[:]) == "stop\n" {
				cancel(errDevReplaced)
				_, _ = io.WriteString(conn, "stopping\n")
			}
			_ = conn.Close()
		}
	}()
	var once sync.Once
	return devCtx, func() {
		once.Do(func() {
			cancel(context.Canceled)
			_ = listener.Close()
			<-done
			unlock()
		})
	}, nil
}
