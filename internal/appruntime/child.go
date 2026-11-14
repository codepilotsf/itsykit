// Package appruntime implements the child side of managed app transport.
package appruntime

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	contactapp "itsy/apps/contact"
	"itsy/internal/appapi"
)

const CapabilityPath = "/_flea/capabilities/contact-email"

// launchConfig is delivered through an inherited pipe, never argv or environment.
type launchConfig struct{ Name, Token, Storage, CapabilityURL, CapabilityToken string }

// RunContact uses fd 3 for the parent's bound listener, fd 4 for credentials,
// and fd 5 as a parent-liveness pipe. It never reads provider environment values.
func RunContact(ctx context.Context) error {
	configFile := os.NewFile(4, "app-credentials")
	var cfg launchConfig
	d := json.NewDecoder(io.LimitReader(configFile, 4097))
	d.DisallowUnknownFields()
	err := d.Decode(&cfg)
	var extra any
	end := d.Decode(&extra)
	configFile.Close()
	if err != nil || end != io.EOF || len(cfg.Token) != 64 || len(cfg.CapabilityToken) != 64 {
		return fmt.Errorf("invalid app launch configuration")
	}
	file := os.NewFile(3, "app-listener")
	listener, err := net.FileListener(file)
	file.Close()
	if err != nil {
		return fmt.Errorf("app listener is unavailable")
	}
	defer listener.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alive := os.NewFile(5, "app-parent")
	defer alive.Close()
	go func() { _, _ = io.Copy(io.Discard, alive); cancel() }()
	h := contactapp.Handler{Notify: func(ctx context.Context, s contactapp.Submission) error {
		return notify(ctx, cfg, s)
	}}
	server := &http.Server{ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "Bearer " + cfg.Token
		if len(r.Header.Values("Authorization")) != 1 || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte(want)) != 1 {
			http.Error(w, "Unauthorized", 401)
			return
		}
		if r.URL.Path == "/_flea/handshake" && r.URL.RawPath == "" && r.URL.RawQuery == "" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(appapi.Identity{Name: "contact", Version: contactapp.Version, Protocol: appapi.Version})
			return
		}
		h.ServeHTTP(w, r)
	})}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
	}
	stop, finish := context.WithTimeout(context.Background(), 2*time.Second)
	defer finish()
	if err := server.Shutdown(stop); err != nil {
		server.Close()
	}
	<-done
	return nil
}
