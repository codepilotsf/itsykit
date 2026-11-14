// A Linux-only, opt-in capacity probe. Run against a separate server executable
// inside an externally resource-limited cgroup; never use installed site configs.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type sample struct{ RSSKiB, FDs int }
type result struct {
	Sites                       int
	AssetMiB                    int
	StartupMS                   int64
	Idle                        sample
	Peak                        sample
	PeakCombinedRSSKiB          int
	Requests, Failures, Skipped int
	P50MS, P95MS, P99MS         float64
	ValidationMS                int64
}

func main() {
	server := flag.String("server", "", "absolute path to isolated flea-server executable")
	root := flag.String("root", "", "existing private test directory (new fixtures subdirectory required)")
	flag.Parse()
	if err := run(*server, *root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func write(name string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	return os.WriteFile(name, b, 0600)
}

func run(server, root string) error {
	if !filepath.IsAbs(server) || !filepath.IsAbs(root) {
		return fmt.Errorf("server and root must be absolute")
	}
	base := filepath.Join(root, "fixtures")
	if err := os.Mkdir(base, 0700); err != nil {
		return err
	}
	// Ordinary regular files, unique inodes per site. Not sparse or hard-linked.
	asset := make([]byte, 1<<20)
	for i := range asset {
		asset[i] = byte(i % 251)
	}
	for n := 0; n < 30; n++ {
		dir := filepath.Join(base, fmt.Sprintf("site-%02d", n))
		config := fmt.Sprintf(`{"app":{"site_name":"Capacity site %02d"}}`, n)
		if err := write(filepath.Join(dir, "itsy.config.json"), []byte(config)); err != nil {
			return err
		}
		items := make([]map[string]string, 20)
		for i := range items {
			items[i] = map[string]string{"title": fmt.Sprintf("Article %d", i), "body": strings.Repeat("Sample website content. ", 12)}
		}
		data, _ := json.Marshal(map[string]any{"items": items})
		if err := write(filepath.Join(dir, "itsy/data/articles.json"), data); err != nil {
			return err
		}
		for i := 0; i < 5; i++ {
			name := fmt.Sprintf("page-%d.html", i)
			if i == 0 {
				name = "index.html"
			}
			page := `<!doctype html><html><head><title>{app.site_name}</title></head><body><h1>{app.site_name}</h1>{#each data.articles.items as item}<article><h2>{item.title}</h2><p>{item.body}</p></article>{/each}</body></html>`
			if err := write(filepath.Join(dir, "site", name), []byte(page)); err != nil {
				return err
			}
		}
		for i := 0; i < 20; i++ {
			if err := write(filepath.Join(dir, "site/assets", fmt.Sprintf("asset-%02d.bin", i)), asset); err != nil {
				return err
			}
		}
	}
	for _, count := range []int{1, 10, 30} {
		res, err := measure(server, root, base, count)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(res)
		fmt.Println(string(b))
	}
	return nil
}

func proc(pid int) sample {
	b, _ := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	s := sample{}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			f := strings.Fields(line)
			s.RSSKiB, _ = strconv.Atoi(f[1])
		}
	}
	fds, _ := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	s.FDs = len(fds)
	return s
}

func measure(server, root, base string, count int) (out result, err error) {
	out.Sites, out.AssetMiB = count, count*20
	sites := map[string]string{}
	for i := 0; i < count; i++ {
		sites[fmt.Sprintf("site-%02d.test", i)] = filepath.Join(base, fmt.Sprintf("site-%02d", i))
	}
	b, _ := json.Marshal(map[string]any{"listen": "127.0.0.1:18089", "sites": sites})
	config := filepath.Join(root, fmt.Sprintf("server-%d.json", count))
	if err = write(config, b); err != nil {
		return
	}
	log, e := os.Create(filepath.Join(root, fmt.Sprintf("server-%d.log", count)))
	if e != nil {
		return out, e
	}
	defer log.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, server, "-config", config)
	cmd.Env = []string{"GOMEMLIMIT=192MiB", "GOMAXPROCS=1"}
	cmd.Stdout, cmd.Stderr = log, log
	start := time.Now()
	if err = cmd.Start(); err != nil {
		return
	}
	defer func() { _ = cmd.Process.Signal(syscall.SIGTERM); _ = cmd.Wait() }()
	transport := &http.Transport{MaxIdleConnsPerHost: 16, MaxConnsPerHost: 16}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 3 * time.Second}
	for {
		resp, e := client.Get("http://127.0.0.1:18089/_flea/health")
		if e == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode == 200 {
				break
			}
		}
		if time.Since(start) > 20*time.Second {
			return out, fmt.Errorf("%d sites: startup failed; see isolated log", count)
		}
		time.Sleep(50 * time.Millisecond)
	}
	out.StartupMS = time.Since(start).Milliseconds()
	time.Sleep(time.Second)
	out.Idle = proc(cmd.Process.Pid)
	var mu sync.Mutex
	var latencies []float64
	stopSample := make(chan struct{})
	doneSample := make(chan struct{})
	validationPID := 0
	go func() {
		defer close(doneSample)
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopSample:
				return
			case <-ticker.C:
				s := proc(cmd.Process.Pid)
				mu.Lock()
				out.Peak.RSSKiB = max(out.Peak.RSSKiB, s.RSSKiB)
				out.Peak.FDs = max(out.Peak.FDs, s.FDs)
				combined := s.RSSKiB
				if validationPID != 0 {
					combined += proc(validationPID).RSSKiB
				}
				out.PeakCombinedRSSKiB = max(out.PeakCombinedRSSKiB, combined)
				mu.Unlock()
			}
		}
	}()
	finishSample := sync.OnceFunc(func() { close(stopSample); <-doneSample })
	defer finishSample()
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	tick := time.NewTicker(time.Second / 60)
	defer tick.Stop()
	var validator *exec.Cmd
	var validateStart time.Time
	for i := 0; i < 900; i++ {
		<-tick.C
		if i == 450 {
			validator = exec.CommandContext(ctx, server, "-config", config, "-validate")
			validator.Env = cmd.Env
			validator.Stdout, validator.Stderr = log, log
			validateStart = time.Now()
			if e := validator.Start(); e != nil {
				wg.Wait()
				finishSample()
				return out, e
			}
			mu.Lock()
			validationPID = validator.Process.Pid
			mu.Unlock()
			wg.Add(1)
			go func() {
				defer wg.Done()
				e := validator.Wait()
				mu.Lock()
				defer mu.Unlock()
				out.ValidationMS = time.Since(validateStart).Milliseconds()
				if e != nil {
					out.Failures++
				}
				validationPID = 0
			}()
		}
		select {
		case sem <- struct{}{}:
		default:
			mu.Lock()
			out.Skipped++
			mu.Unlock()
			continue
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			// Rotate across sites independently of request kind: all sites get assets.
			site := (i / 5) % count
			path := "/"
			if i%5 == 4 {
				path = fmt.Sprintf("/assets/asset-%02d.bin", (i/5)%20)
			}
			req, _ := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:18089"+path, nil)
			req.Host = fmt.Sprintf("site-%02d.test", site)
			started := time.Now()
			resp, e := client.Do(req)
			ok := e == nil
			if e == nil {
				if path == "/" {
					body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
					ok = readErr == nil && strings.Contains(string(body), fmt.Sprintf("Capacity site %02d", site))
				} else {
					n, readErr := io.Copy(io.Discard, resp.Body)
					ok = readErr == nil && n == 1<<20
				}
				resp.Body.Close()
				ok = ok && resp.StatusCode == 200
			}
			mu.Lock()
			defer mu.Unlock()
			out.Requests++
			if !ok {
				out.Failures++
			}
			latencies = append(latencies, float64(time.Since(started).Microseconds())/1000)
		}(i)
	}
	wg.Wait()
	finishSample()
	mu.Lock()
	sort.Float64s(latencies)
	if len(latencies) > 0 {
		out.P50MS = latencies[len(latencies)*50/100]
		out.P95MS = latencies[len(latencies)*95/100]
		out.P99MS = latencies[len(latencies)*99/100]
	}
	mu.Unlock()
	return out, nil
}
