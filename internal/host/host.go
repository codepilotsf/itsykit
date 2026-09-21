// Package host runs the persistent server, without development reload scripts,
// filesystem watching, or network access to resolve icons.
package host

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nanoweb/internal/site"
)

const HealthPath = "/_nanoweb/health"

type Config struct {
	Listen string            `json:"listen"`
	Sites  map[string]string `json:"sites"`
}

func Load(configPath string, logger *log.Logger) (Config, http.Handler, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, nil, err
	}
	var config Config
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return config, nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return config, nil, fmt.Errorf("server config must contain one JSON object")
	}
	address, port, err := net.SplitHostPort(config.Listen)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || portNumber < 1 || portNumber > 65535 || (address != "" && net.ParseIP(address) == nil) {
		return config, nil, fmt.Errorf("server listen must be an IP address (or empty wildcard) and numeric port 1-65535")
	}
	handlers := map[string]http.Handler{}
	for host, root := range config.Sites {
		if host == "" || host != strings.ToLower(host) || strings.ContainsAny(host, " /\\\r\n\t:") || strings.HasSuffix(host, ".") {
			return config, nil, fmt.Errorf("invalid site hostname %q", host)
		}
		if !filepath.IsAbs(root) {
			return config, nil, fmt.Errorf("site %s: root must be absolute", host)
		}
		project, err := site.Load(root)
		if err != nil {
			return config, nil, fmt.Errorf("site %s: %w", host, err)
		}
		if logger != nil {
			for _, warning := range project.Warnings {
				logger.Printf("Warning: %s: %s", host, warning)
			}
		}
		handlers[host] = site.NewProductionServer(project, logger)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == HealthPath {
			w.Header().Set("Content-Type", "application/json")
			if r.Method != "GET" && r.Method != "HEAD" {
				w.Header().Set("Allow", "GET, HEAD")
				w.WriteHeader(405)
				return
			}
			if r.Method == "GET" {
				_, _ = io.WriteString(w, `{"service":"nanoweb","status":"ready"}`)
			}
			return
		}
		host := r.Host
		if name, _, err := net.SplitHostPort(host); err == nil {
			host = name
		}
		host = strings.TrimSuffix(strings.ToLower(host), ".")
		if handler, ok := handlers[host]; ok {
			handler.ServeHTTP(w, r)
			return
		}
		if logger != nil {
			logger.Printf("404 %s %s", r.Method, r.URL.EscapedPath())
		}
		http.NotFound(w, r)
	})
	return config, handler, nil
}

func Run(ctx context.Context, configPath string, logger *log.Logger) error {
	config, handler, err := Load(configPath, logger)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return err
	}
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	if logger != nil {
		logger.Printf("Nanoweb server listening on %s", listener.Addr())
	}
	select {
	case err := <-done:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
		}
		<-done
		return nil
	}
}

func Check(ctx context.Context, configPath string) error {
	config, _, err := Load(configPath, nil)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, "GET", "http://"+checkAddress(config.Listen)+HealthPath, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1024))
	if err != nil {
		return err
	}
	if response.StatusCode != 200 || string(body) != `{"service":"nanoweb","status":"ready"}` {
		return fmt.Errorf("Nanoweb server readiness check failed")
	}
	return nil
}

// Probe wildcard listeners through loopback, never an unspecified destination.
func checkAddress(listen string) string {
	address, port, _ := net.SplitHostPort(listen)
	switch address {
	case "", "0.0.0.0":
		address = "127.0.0.1"
	case "::":
		address = "::1"
	}
	return net.JoinHostPort(address, port)
}
