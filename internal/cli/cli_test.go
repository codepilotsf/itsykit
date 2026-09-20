package cli

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"nanoweb/internal/site"
)

func TestCLI(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{nil, {"unknown"}, {"new"}, {"new", "../escape"}, {"dev", "-port", "-1"}, {"status", "extra"}} {
		if code := Run(context.Background(), args, io.Discard, io.Discard); code != 2 {
			t.Fatalf("%v returned %d", args, code)
		}
	}
	if code := Run(context.Background(), []string{"status"}, io.Discard, io.Discard); code != 1 {
		t.Fatal(code)
	}
	if code := Run(context.Background(), []string{"new", "demo"}, io.Discard, io.Discard); code != 0 {
		t.Fatal(code)
	}
	before, err := os.ReadFile("demo/html/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if code := Run(context.Background(), []string{"new", "demo"}, io.Discard, io.Discard); code != 1 {
		t.Fatal("overwrote existing project")
	}
	after, err := os.ReadFile("demo/html/index.html")
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("existing content changed")
	}
	if _, err := site.Load("demo"); err != nil {
		t.Fatal(err)
	}
	t.Chdir("demo")
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"status"}, &out, io.Discard); code != 0 || !strings.Contains(out.String(), "Valid local project") {
		t.Fatal(code, out.String())
	}
	if err := os.WriteFile("html/index.html", []byte("{broken syntax}"), 0644); err != nil {
		t.Fatal(err)
	}
	if code := Run(context.Background(), []string{"dev"}, io.Discard, io.Discard); code != 1 {
		t.Fatal("dev did not validate first")
	}
}

// This is an actual executable smoke test: scaffold, status, local server,
// HTTP requests, edits, invalid state, recovery, and graceful shutdown.
func TestExecutableWorkflow(t *testing.T) {
	temp := t.TempDir()
	binary := filepath.Join(temp, "nanoweb")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "../../cmd/nanoweb")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	run := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command(binary, args...)
		cmd.Dir = dir
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, b)
		}
		return string(b)
	}
	run(temp, "new", "smoke")
	root := filepath.Join(temp, "smoke")
	if got := run(root, "status"); !strings.Contains(got, "2 pages, 2 assets, 1 snips") {
		t.Fatal(got)
	}
	cmd := exec.Command(binary, "dev", "-port", "0")
	cmd.Dir = root
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	stopped := false
	defer func() {
		if !stopped {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()
	lines := make(chan string, 8)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	var address string
	select {
	case line, ok := <-lines:
		if !ok {
			err := cmd.Wait()
			stopped = true
			t.Fatalf("dev exited: %v: %s", err, stderr.String())
		}
		address = strings.TrimPrefix(line, "Nanoweb running at ")
	case <-time.After(10 * time.Second):
		t.Fatal("dev did not start")
	}
	client := &http.Client{Timeout: 2 * time.Second}
	get := func(url string) (int, string) {
		t.Helper()
		response, err := client.Get(address + url)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		b, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, string(b)
	}
	for url, want := range map[string]int{"/": 200, "/about": 200, "/css/style.css": 200, "/images/mark.svg": 200, "/about/": 404, "/about.html": 404, "/_/nanoweb.config": 404, "/_header.html": 404, "/%2e%2e/_/nanoweb.config": 404} {
		if code, _ := get(url); code != want {
			t.Errorf("%s: %d", url, code)
		}
	}
	edit := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	wait := func(url string, want int, text string) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			code, body := get(url)
			if code == want && strings.Contains(body, text) {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		t.Fatalf("%s never reached %d containing %q", url, want, text)
	}
	// Ensure the watcher has begun before the first edit.
	time.Sleep(100 * time.Millisecond)
	edit("_/snips/home.json", `{"title":"Smoke test changed","services":[]}`)
	wait("/", 200, "Smoke test changed")
	edit("html/css/style.css", "/* updated asset */")
	wait("/css/style.css", 200, "updated asset")
	edit("_/nanoweb.config", `{"site":{"name":"Changed config"}}`)
	wait("/", 200, "Changed config")
	edit("html/index.html", `{render missing-secret-name}`)
	wait("/", 500, "temporarily unavailable")
	if _, body := get("/"); strings.Contains(body, "missing-secret-name") {
		t.Fatal("error detail leaked")
	}
	invalid := exec.Command(binary, "status")
	invalid.Dir = root
	if err := invalid.Run(); err == nil {
		t.Fatal("invalid status exited successfully")
	}
	edit("html/index.html", "<h1>Recovered template</h1>")
	wait("/", 200, "Recovered template")
	run(root, "status")
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		stopped = true
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("dev did not shut down")
	}
	if !strings.Contains(stderr.String(), "missing partial") || !strings.Contains(stderr.String(), "Site valid again") {
		t.Fatalf("missing error/recovery diagnostics: %s", stderr.String())
	}
	t.Log("Executable smoke passed: scaffold, validation, HTTP, private paths, snip/config/asset/template reloads, invalid state, recovery, shutdown")
}
