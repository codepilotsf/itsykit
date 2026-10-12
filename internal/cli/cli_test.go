package cli

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"flea/internal/site"
	"flea/templates"
)

func TestCreateCopiesEmbeddedTemplates(t *testing.T) {
	root := filepath.Join(t.TempDir(), "site")
	if err := createSite(root, true); err != nil {
		t.Fatal(err)
	}
	err := fs.WalkDir(templates.Files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		want, err := templates.Files.ReadFile(name)
		if err != nil {
			return err
		}
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		if !bytes.Equal(got, want) {
			t.Errorf("generated %s differs from embedded template", name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".flea.config.json", "flea/snips.json", "site/main.layout.html", "site/_404.html", "site/_500.html"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "embed.go")); !os.IsNotExist(err) {
		t.Fatal("embedding code must not be scaffolded")
	}
}

func TestHelpWithoutSite(t *testing.T) {
	if !strings.HasSuffix(usage, "\n\n") {
		t.Fatal("overview needs trailing blank line")
	}
	t.Chdir(t.TempDir())
	for _, args := range [][]string{nil, {"help"}, {"--help"}, {"-h"}} {
		var out, errOut bytes.Buffer
		if code := Run(context.Background(), args, &out, &errOut); code != 0 || errOut.Len() != 0 || out.String() != usage {
			t.Fatalf("%v: code=%d stdout=%q stderr=%q", args, code, out.String(), errOut.String())
		}
	}
	for _, command := range []string{"new", "dev", "status", "admin", "email", "setup", "deploy", "https", "version", "pull", "update", "update-server", "server"} {
		for _, flag := range []string{"--help", "-h"} {
			var out, errOut bytes.Buffer
			if code := Run(context.Background(), []string{command, flag}, &out, &errOut); code != 0 || errOut.Len() != 0 || !strings.Contains(out.String(), "Usage: flea "+command) || !strings.HasSuffix(out.String(), "\n\n") {
				t.Fatalf("%s %s: code=%d stdout=%q stderr=%q", command, flag, code, out.String(), errOut.String())
			}
		}
	}
}

func TestCLI(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{{"unknown"}, {"upgrade"}, {"update", "extra"}, {"new"}, {"new", "../escape"}, {"dev", "-port", "-1"}, {"status", "extra"}, {"setup", "extra"}, {"deploy", "extra"}, {"https", "extra"}} {
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
	before, err := os.ReadFile("demo/site/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if code := Run(context.Background(), []string{"new", "demo"}, io.Discard, io.Discard); code != 1 {
		t.Fatal("overwrote existing project")
	}
	after, err := os.ReadFile("demo/site/index.html")
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("existing content changed")
	}
	if _, err := site.Load("demo"); err != nil {
		t.Fatal(err)
	}
	t.Chdir("demo")
	var out bytes.Buffer
	if code := Run(context.Background(), []string{"status"}, &out, io.Discard); code != 0 || !strings.Contains(out.String(), "Valid configuration, templates and content") {
		t.Fatal(code, out.String())
	}
	if err := os.WriteFile("site/index.html", []byte("{broken syntax}"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var devErr bytes.Buffer
	if code := Run(ctx, []string{"dev", "-port", "0"}, io.Discard, &devErr); code != 1 || !strings.Contains(devErr.String(), "index.html") {
		t.Fatal("dev did not validate the page first", code, devErr.String())
	}
}

func TestStatusWarnsForMissingIcon(t *testing.T) {
	root := filepath.Join(t.TempDir(), "site")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "site", "index.html"), []byte(`Hello{icon unknown-icon}`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	var out, warnings bytes.Buffer
	if code := Run(context.Background(), []string{"status"}, &out, &warnings); code != 0 {
		t.Fatal(code, warnings.String())
	}
	if !strings.Contains(warnings.String(), "Warning: index.html:1:") || !strings.Contains(warnings.String(), "icon omitted") {
		t.Fatal(warnings.String())
	}
	if _, err := os.Stat("flea/icons/unknown-icon.svg"); !os.IsNotExist(err) {
		t.Fatal("status wrote cache")
	}
}

// This is an actual executable smoke test: scaffold, status, local server,
// HTTP requests, edits, invalid state, recovery, and graceful shutdown.
func TestExecutableWorkflow(t *testing.T) {
	temp := t.TempDir()
	binary := filepath.Join(temp, "flea")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "../../cmd/flea")
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
	run(temp, "new", "--starter", "smoke")
	root := filepath.Join(temp, "smoke")
	if got := run(root, "status"); !strings.Contains(got, "Smoke\n") {
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
	startupTimeout := time.After(10 * time.Second)
	for address == "" {
		select {
		case line, ok := <-lines:
			if !ok {
				err := cmd.Wait()
				stopped = true
				t.Fatalf("dev exited: %v: %s", err, stderr.String())
			}
			if strings.HasPrefix(line, "Website  ") {
				address = strings.TrimPrefix(line, "Website  ")
			}
		case <-startupTimeout:
			t.Fatal("dev did not start")
		}
	}
	if !strings.HasPrefix(address, "http://localhost:") || address == "http://localhost:0" {
		t.Fatalf("expected localhost with the assigned port, got %q", address)
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
	for url, want := range map[string]int{"/": 200, "/about": 404, "/css/global.css": 200, "/images/mark.svg": 200, "/about/": 404, "/about.html": 404, "/.flea.config.json": 404, "/_header.html": 404, "/%2e%2e/.flea.config.json": 404} {
		if code, _ := get(url); code != want {
			t.Errorf("%s: %d", url, code)
		}
	}
	for _, name := range []string{"theme.css", "global.css", "typography.css", "forms.css", "ui.css"} {
		if code, body := get("/css/" + name); code != 200 || !strings.Contains(body, "{") {
			t.Errorf("starter CSS %s: status %d, body %q", name, code, body)
		}
	}
	edit := func(name, body string) {
		t.Helper()
		if strings.HasSuffix(name, ".html") || strings.HasPrefix(name, "css/") {
			name = "site/" + name
		}
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
	seedContent, err := os.ReadFile(filepath.Join(root, "flea", "snips.json"))
	if err != nil {
		t.Fatal(err)
	}
	edit("flea/snips.json", strings.Replace(string(seedContent), "Welcome to FleaCMS.", "Smoke test changed", 1))
	wait("/", 200, "Smoke test changed")
	edit("test.html", `<section snip="intro"><p>Extracted <em>live</em></p></section>`)
	wait("/test", 200, "Extracted <em>live</em>")
	extracted, err := os.ReadFile(filepath.Join(root, "site", "test.html"))
	if err != nil || !strings.Contains(string(extracted), `{snip `) || strings.Contains(string(extracted), `{snip "`) {
		t.Fatalf("region was not extracted: %s %v", extracted, err)
	}
	// Removal is reconciled on save, and undo restores the archived content.
	edit("test.html", `<h1>Region removed</h1>`)
	wait("/test", 200, "Region removed")
	saved, err := os.ReadFile(filepath.Join(root, "flea", "snips.json"))
	if err != nil || !strings.Contains(string(saved), `"archive"`) {
		t.Fatal("missing archive after removal", err)
	}
	edit("test.html", string(extracted))
	wait("/test", 200, "Extracted <em>live</em>")
	if _, err := os.Stat(filepath.Join(root, "flea", ".snips.lock")); !os.IsNotExist(err) {
		t.Fatal("dev created lock file", err)
	}
	edit("css/global.css", "/* updated asset */")
	wait("/css/global.css", 200, "updated asset")
	edit(".flea.config.json", `{"app":{"site_name":"Changed config"}}`)
	wait("/", 200, "Changed config")
	edit("index.html", `{render missing-secret-name}`)
	wait("/", 500, "")
	if _, body := get("/"); !strings.Contains(body, "How to fix it") || !strings.Contains(body, "missing-secret-name") {
		t.Fatal("local dev error did not explain the missing partial")
	}
	invalid := exec.Command(binary, "status")
	invalid.Dir = root
	if err := invalid.Run(); err == nil {
		t.Fatal("invalid status exited successfully")
	}
	edit("index.html", "<h1>Recovered template</h1>")
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
	for _, line := range []string{"GET /\n", "404 GET /about\n", "500 GET /\n", "Snips updated in flea/snips.json.\n"} {
		if !strings.Contains(stderr.String(), line) {
			t.Errorf("missing request log %q", line)
		}
	}
	t.Log("Executable smoke passed: scaffold, validation, HTTP, private paths, flea/config/asset/template reloads, invalid state, recovery, shutdown")
}
