// The inline editor is a standalone, local-only experiment, not a Flea route.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"flea/internal/snips"
	"flea/templates"
)

//go:embed index.html prototype.css app.js checks.js vendor/*
var files embed.FS

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(files)))
	mux.Handle("/site/", http.StripPrefix("/site/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Clean(r.URL.Path) != r.URL.Path || (!strings.HasPrefix(r.URL.Path, "css/") && r.URL.Path != "images/mark.svg") {
			http.NotFound(w, r)
			return
		}
		http.FileServer(http.FS(templates.Files)).ServeHTTP(w, r)
	})))
	mux.HandleFunc("/sanitize-check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		if r.Header.Get("Origin") != "http://"+r.Host {
			w.WriteHeader(403)
			return
		}
		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "Content too large", 413)
			return
		}
		clean, err := snips.Sanitize(string(data))
		if err != nil {
			http.Error(w, "Invalid content", 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"html": clean, "changed": clean != string(data)})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil || host != "127.0.0.1" {
			http.Error(w, "Use the printed loopback URL.", 403)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self' https://fonts.googleapis.com 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com; img-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		mux.ServeHTTP(w, r)
	})
}

func main() {
	port := flag.Int("port", 3143, "loopback preview port")
	flag.Parse()
	server := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", *port), Handler: handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	log.Printf("Inline-editor prototype: http://%s (demo content only; saves stay in this tab)", server.Addr)
	log.Fatal(server.ListenAndServe())
}
