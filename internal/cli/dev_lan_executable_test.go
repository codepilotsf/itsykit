package cli

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"flea/internal/site"
)

func TestLANExecutableFlags(t *testing.T) {
	addresses, err := devLANAddresses()
	if err != nil {
		t.Skip("requires a private IPv4 interface: ", err)
	}
	binary := filepath.Join(t.TempDir(), "itsy")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", binary, "../../cmd/flea")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	for _, flag := range []string{"", "--lan", "-l"} {
		t.Run(map[bool]string{true: "default", false: flag}[flag == ""], func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "site")
			if err := Create(root); err != nil {
				t.Fatal(err)
			}
			if _, _, err := site.AddAdmin(root, "owner@example.test"); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			args := []string{"dev", "--port", "0"}
			if flag != "" {
				args = append(args, flag)
			}
			cmd := exec.CommandContext(ctx, binary, args...)
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
			defer func() { _ = cmd.Process.Signal(os.Interrupt); _ = cmd.Wait() }()
			lines := make(chan string, 32)
			go func() {
				defer close(lines)
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					lines <- scanner.Text()
				}
			}()
			var website, network, hostname string
			ready := false
			for !ready {
				select {
				case line, ok := <-lines:
					if !ok {
						t.Fatal("dev exited before startup")
					}
					if strings.HasPrefix(line, "Website  ") {
						website = strings.TrimPrefix(line, "Website  ")
					}
					if strings.HasPrefix(line, "Network  ") && network == "" {
						network = strings.TrimPrefix(line, "Network  ")
					}
					if strings.HasPrefix(line, "Hostname ") && hostname == "" {
						hostname = strings.TrimPrefix(line, "Hostname ")
					}
					ready = strings.Contains(line, "Press Ctrl-C")
				case <-ctx.Done():
					t.Fatal("dev startup timeout")
				}
			}
			transport := &http.Transport{Proxy: nil}
			defer transport.CloseIdleConnections()
			client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
			get := func(base, path string) {
				t.Helper()
				response, err := client.Get(base + path)
				if err != nil {
					t.Fatal(err)
				}
				defer response.Body.Close()
				_, _ = io.Copy(io.Discard, response.Body)
				if response.StatusCode != 200 {
					t.Fatalf("%s: %d", path, response.StatusCode)
				}
			}
			get(website, "/")
			if flag == "" {
				if network != "" {
					t.Fatal("default advertised LAN")
				}
				response, err := client.Get(strings.Replace(website, "localhost", addresses[0], 1))
				if err == nil {
					response.Body.Close()
					t.Fatal("default dev accepted LAN connection")
				}
				return
			}
			if network == "" {
				t.Fatal("missing network URL")
			}
			get(network, "/")
			get(network, "/admin")
			get(network, "/_flea/reload")
			if hostname == "" {
				t.Fatal("missing hostname URL")
			}
			alias, err := url.Parse(hostname)
			if err != nil {
				t.Fatal(err)
			}
			aliasRequest, _ := http.NewRequest("GET", network+"/admin", nil)
			aliasRequest.Host = alias.Host
			aliasResponse, err := client.Do(aliasRequest)
			if err != nil {
				t.Fatal(err)
			}
			aliasResponse.Body.Close()
			if aliasResponse.StatusCode != 200 {
				t.Fatal("advertised hostname blocked", aliasResponse.Status)
			}
			r, _ := http.NewRequest("GET", network+"/", nil)
			r.Host = "evil.test:3000"
			response, err := client.Do(r)
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != 403 {
				t.Fatal("unapproved Host accepted")
			}
		})
	}
}
