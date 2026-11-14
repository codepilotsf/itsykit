// Command session-app is an independent, standard-library example of a custom
// Flea app that owns its own server-side sessions. It is not part of Flea and
// imports no Flea package. Copy this directory into a site, build the binary for
// the server platform, and register it under apps.session.
//
// Flea does not run app business logic or store app sessions. It launches this
// executable, forwards checked form fields as flat JSON, carries the browser's
// opaque token as the X-Flea-Session header, and turns the response's `session`
// value into its own host-only cookie. The token never appears in templates,
// logs, URLs or public JSON.
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	protocol      = 2
	version       = "0.1.0"
	sessionHeader = "X-Flea-Session"
	defaultTTL    = 3600
	maxSessions   = 1000
)

type launchConfig struct {
	Name            string
	Token           string
	Storage         string
	CapabilityURL   string
	CapabilityToken string
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	configFile := os.NewFile(4, "app-credentials")
	var cfg launchConfig
	decoder := json.NewDecoder(io.LimitReader(configFile, 4097))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&cfg)
	var extra any
	end := decoder.Decode(&extra)
	configFile.Close()
	if err != nil || end != io.EOF || !validToken(cfg.Token) {
		return errors.New("invalid launch configuration")
	}
	name := cfg.Name
	if name == "" {
		name = "session"
	}

	listenerFile := os.NewFile(3, "app-listener")
	listener, err := net.FileListener(listenerFile)
	listenerFile.Close()
	if err != nil {
		return errors.New("app listener is unavailable")
	}
	defer listener.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	alive := os.NewFile(5, "app-parent")
	defer alive.Close()
	go func() { _, _ = io.Copy(io.Discard, alive); cancel() }()

	store, err := openStore(cfg.Storage)
	if err != nil {
		return err
	}
	handler := &app{name: name, token: cfg.Token, store: store}
	server := &http.Server{
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 << 10,
		Handler:           handler,
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
	}
	shutdown, finish := context.WithTimeout(context.Background(), 2*time.Second)
	defer finish()
	if err := server.Shutdown(shutdown); err != nil {
		server.Close()
	}
	<-done
	return nil
}

func validToken(token string) bool {
	if len(token) < 16 || len(token) > 512 {
		return false
	}
	for i := 0; i < len(token); i++ {
		if token[i] < 0x21 || token[i] > 0x7e {
			return false
		}
	}
	return true
}

// store is the app's own session database: a small JSON file under the private
// storage directory Flea supplies. It survives app restarts and is never held by
// Flea. This example is intentionally single-process and file-locked only by its
// own mutex. Persistence failures are returned to the caller and fail the
// request closed rather than silently resurrecting revoked sessions.
//
// NOTE: the "login" action is a demonstration only. It performs no credential
// check and must never be treated as real authentication.
type store struct {
	mu     sync.Mutex
	path   string
	tokens map[string]int64 // token -> unix expiry
}

func openStore(dir string) (*store, error) {
	s := &store{tokens: map[string]int64{}}
	if dir == "" {
		return s, nil
	}
	s.path = filepath.Join(dir, "sessions.json")
	data, err := os.ReadFile(s.path)
	if err == nil {
		if err := json.Unmarshal(data, &s.tokens); err != nil {
			return nil, fmt.Errorf("session store is unreadable")
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("session store is unreadable")
	}
	s.expire()
	return s, nil
}

func (s *store) expire() {
	now := time.Now().Unix()
	for token, expiry := range s.tokens {
		if expiry <= now {
			delete(s.tokens, token)
		}
	}
}

// save writes the store durably: a same-directory temp file, fsync, rename and
// directory fsync. Any failure is returned so the request can fail closed.
func (s *store) save() error {
	if s.path == "" {
		return nil
	}
	data, err := json.Marshal(s.tokens)
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	f, err := os.CreateTemp(dir, ".sessions-")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(temp, s.path); err != nil {
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func (s *store) create(ttl int) (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expire()
	if len(s.tokens) >= maxSessions {
		return "", fmt.Errorf("session store is full")
	}
	s.tokens[token] = time.Now().Add(time.Duration(ttl) * time.Second).Unix()
	if err := s.save(); err != nil {
		delete(s.tokens, token)
		return "", err
	}
	return token, nil
}

func (s *store) valid(token string) bool {
	if token == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expire()
	_, ok := s.tokens[token]
	return ok
}

func (s *store) drop(token string) error {
	if token == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	previous, ok := s.tokens[token]
	if !ok {
		return nil
	}
	delete(s.tokens, token)
	if err := s.save(); err != nil {
		s.tokens[token] = previous
		return err
	}
	return nil
}

type app struct {
	name  string
	token string
	store *store
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.Header.Values("Authorization")) != 1 ||
		subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+a.token)) != 1 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.URL.RawPath != "" || r.URL.RawQuery != "" {
		writeJSON(w, 400, map[string]any{"error": "Unsupported request."})
		return
	}
	if r.URL.Path == "/_flea/handshake" && r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]any{"name": a.name, "version": version, "protocol": protocol})
		return
	}
	if r.URL.Path != "/" {
		writeJSON(w, 404, map[string]any{"error": "Unsupported request."})
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.writeState(w, r, http.StatusOK)
	case http.MethodPost:
		a.handlePost(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, 405, map[string]any{"error": "Use GET or POST."})
	}
}

func (a *app) writeState(w http.ResponseWriter, r *http.Request, status int) {
	token := r.Header.Get(sessionHeader)
	body := map[string]any{"data": map[string]any{"authenticated": a.store.valid(token)}}
	if token != "" && !a.store.valid(token) {
		// An expired or revoked token is cleared by omission-safe null.
		body["session"] = nil
	}
	writeJSON(w, status, body)
}

func (a *app) handlePost(w http.ResponseWriter, r *http.Request) {
	fields, err := readFields(w, r)
	if err != nil {
		writeJSON(w, 400, map[string]any{"error": "The request is invalid."})
		return
	}
	current := r.Header.Get(sessionHeader)
	action, _ := fields["action"].(string)
	invalid := map[string]string{}
	switch action {
	case "login", "login-short":
		ttl := defaultTTL
		if action == "login-short" {
			ttl = 1
		}
		token, err := a.store.create(ttl)
		if err != nil {
			writeJSON(w, 503, map[string]any{"error": "Could not start a session."})
			return
		}
		writeJSON(w, 200, map[string]any{
			"data":    map[string]any{"authenticated": true},
			"session": map[string]any{"token": token, "max_age": ttl},
		})
	case "rotate":
		if !a.store.valid(current) {
			writeJSON(w, 200, map[string]any{"data": map[string]any{"authenticated": false}, "session": nil})
			return
		}
		if err := a.store.drop(current); err != nil {
			writeJSON(w, 503, map[string]any{"error": "Could not rotate the session."})
			return
		}
		token, err := a.store.create(defaultTTL)
		if err != nil {
			writeJSON(w, 503, map[string]any{"error": "Could not rotate the session."})
			return
		}
		writeJSON(w, 200, map[string]any{
			"data":    map[string]any{"authenticated": true},
			"session": map[string]any{"token": token, "max_age": defaultTTL},
		})
	case "logout":
		if err := a.store.drop(current); err != nil {
			writeJSON(w, 503, map[string]any{"error": "Could not end the session."})
			return
		}
		writeJSON(w, 200, map[string]any{"data": map[string]any{"authenticated": false}, "session": nil})
	default:
		invalid["action"] = "Use login, rotate or logout."
		writeJSON(w, 422, map[string]any{"data": fields, "invalid": invalid})
	}
}

// readFields accepts the flat object of string values Flea sends. Repeated
// values arrive as string arrays; this example rejects them rather than guessing.
func readFields(w http.ResponseWriter, r *http.Request) (map[string]any, error) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return nil, errors.New("expected JSON")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 192<<10)
	var fields map[string]any
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&fields); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("expected one object")
	}
	for name, value := range fields {
		if _, ok := value.(string); !ok {
			return nil, fmt.Errorf("field %q must be a single string", name)
		}
		if name == "_csrf" {
			return nil, errors.New("CSRF belongs to Flea")
		}
	}
	return fields, nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
