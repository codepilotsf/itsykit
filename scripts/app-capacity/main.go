// Linux-only opt-in probe. Requires an externally isolated, resource-capped unit.
// Never reads live configs; all fixtures are new and all signaled PIDs are children.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type usage struct{ RSSKiB, PSSKiB, FDs, Threads int }
type result struct {
	Sites                       int
	StartupMS                   int64
	Parent, Children, Total     usage
	PeakTotal                   usage
	Requests, Failures          int
	P95MS                       float64
	CrashRecoveryMS, ShutdownMS int64
	ParentDeath                 bool
}

func main() {
	server := flag.String("server", "", "isolated server executable")
	root := flag.String("root", "", "fresh fixture parent")
	flag.Parse()
	if err := run(*server, *root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func write(name, body string) error {
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	return os.WriteFile(name, []byte(body), 0600)
}
func run(server, root string) error {
	if runtime.GOOS != "linux" || !filepath.IsAbs(server) || !filepath.IsAbs(root) {
		return fmt.Errorf("Linux and absolute isolated paths required")
	}
	base := filepath.Join(root, "app-fixtures")
	if err := os.Mkdir(base, 0700); err != nil {
		return err
	}
	for n := 0; n < 30; n++ {
		dir := filepath.Join(base, fmt.Sprintf("site-%02d", n))
		files := map[string]string{
			".itsy.config.json":           `{"app":{"site_name":"Contact capacity"},"apps":{"contact":{"mount":"/contact","email":{"to":"owner@example.test"}}}}`,
			"site/index.html":             `<h1>{app.site_name}</h1><p>Ordinary page.</p>`,
			"site/contact/index.html":     `<form method="post" action="/contact">{csrf}<input name="name" value="{contact.data.name}"><p>{contact.error}</p></form>`,
			"site/contact/thank-you.html": `<h1>Thanks</h1><p>{contact.flash.notice}</p>`,
		}
		for name, body := range files {
			if err := write(filepath.Join(dir, name), body); err != nil {
				return err
			}
		}
	}
	for _, count := range []int{1, 10, 30} {
		out, err := measure(server, root, base, count, false)
		if err != nil {
			return err
		}
		if err = json.NewEncoder(os.Stdout).Encode(out); err != nil {
			return err
		}
	}
	out, err := measure(server, root, base, 1, true)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(out)
}
func readUsage(pid int) usage {
	var u usage
	b, _ := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		n, _ := strconv.Atoi(f[1])
		switch f[0] {
		case "VmRSS:":
			u.RSSKiB = n
		case "Threads:":
			u.Threads = n
		}
	}
	b, _ = os.ReadFile(fmt.Sprintf("/proc/%d/smaps_rollup", pid))
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 && f[0] == "Pss:" {
			u.PSSKiB, _ = strconv.Atoi(f[1])
		}
	}
	files, _ := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
	u.FDs = len(files)
	return u
}
func children(pid int) []int {
	// Children can be forked by any Go OS thread, not only the thread-group leader.
	tasks, _ := os.ReadDir(fmt.Sprintf("/proc/%d/task", pid))
	seen := map[int]bool{}
	for _, task := range tasks {
		b, _ := os.ReadFile(fmt.Sprintf("/proc/%d/task/%s/children", pid, task.Name()))
		for _, s := range strings.Fields(string(b)) {
			n, _ := strconv.Atoi(s)
			if n > 0 {
				seen[n] = true
			}
		}
	}
	var out []int
	for n := range seen {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}
func alive(pid int) bool {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "State:") {
			return !strings.Contains(line, "Z") && !strings.Contains(line, "X")
		}
	}
	return false
}
func plus(a, b usage) usage {
	return usage{a.RSSKiB + b.RSSKiB, a.PSSKiB + b.PSSKiB, a.FDs + b.FDs, a.Threads + b.Threads}
}
func peak(a, b usage) usage {
	return usage{max(a.RSSKiB, b.RSSKiB), max(a.PSSKiB, b.PSSKiB), max(a.FDs, b.FDs), max(a.Threads, b.Threads)}
}
func snapshot(pid int) (usage, usage, usage) {
	parent := readUsage(pid)
	var kids usage
	for _, child := range children(pid) {
		kids = plus(kids, readUsage(child))
	}
	return parent, kids, plus(parent, kids)
}
func waitUntil(timeout time.Duration, ok func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if ok() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(50 * time.Millisecond)
	}
}
func measure(server, root, base string, count int, killParent bool) (out result, err error) {
	out.Sites = count
	out.ParentDeath = killParent
	sites := map[string]string{}
	for i := 0; i < count; i++ {
		sites[fmt.Sprintf("site-%02d.test", i)] = filepath.Join(base, fmt.Sprintf("site-%02d", i))
	}
	b, _ := json.Marshal(map[string]any{"listen": "127.0.0.1:18089", "sites": sites})
	config := filepath.Join(root, fmt.Sprintf("apps-%d-%t.json", count, killParent))
	if err = write(config, string(b)); err != nil {
		return
	}
	log, e := os.Create(filepath.Join(root, fmt.Sprintf("apps-%d-%t.log", count, killParent)))
	if e != nil {
		return out, e
	}
	defer log.Close()
	cmd := exec.Command(server, "-config", config)
	cmd.Env = []string{"GOMAXPROCS=1", "GOMEMLIMIT=192MiB"}
	cmd.Stdout, cmd.Stderr = log, log
	started := time.Now()
	if err = cmd.Start(); err != nil {
		return
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	reaped := false
	defer func() {
		if !reaped {
			cmd.Process.Kill()
			<-done
		}
	}()
	client := &http.Client{Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	get := func(host, route string) bool {
		r, _ := http.NewRequest("GET", "http://127.0.0.1:18089"+route, nil)
		r.Host = host
		response, e := client.Do(r)
		if e != nil {
			return false
		}
		defer response.Body.Close()
		_, e = io.Copy(io.Discard, response.Body)
		return e == nil && response.StatusCode == 200
	}
	if !waitUntil(20*time.Second, func() bool { return len(children(cmd.Process.Pid)) == count && get("site-00.test", "/") }) {
		return out, fmt.Errorf("%d sites failed startup; inspect isolated log", count)
	}
	// Counts establish launch, not protocol readiness; native integration tests
	// separately exercise authenticated handshake and actual form submissions.
	out.StartupMS = time.Since(started).Milliseconds()
	time.Sleep(time.Second)
	out.Parent, out.Children, out.Total = snapshot(cmd.Process.Pid)
	out.PeakTotal = out.Total
	latencies := []float64{}
	for i := 0; i < 300; i++ {
		start := time.Now()
		if !get(fmt.Sprintf("site-%02d.test", i%count), "/") {
			out.Failures++
		}
		out.Requests++
		latencies = append(latencies, float64(time.Since(start).Microseconds())/1000)
		_, _, total := snapshot(cmd.Process.Pid)
		out.PeakTotal = peak(out.PeakTotal, total)
		time.Sleep(20 * time.Millisecond)
	}
	sort.Float64s(latencies)
	out.P95MS = latencies[len(latencies)*95/100]
	pids := children(cmd.Process.Pid)
	if len(pids) != count {
		return out, fmt.Errorf("unexpected child count before recovery")
	}
	victim := pids[0]
	if err = syscall.Kill(victim, syscall.SIGKILL); err != nil {
		return
	}
	start := time.Now()
	if !waitUntil(10*time.Second, func() bool {
		kids := children(cmd.Process.Pid)
		if len(kids) != count {
			return false
		}
		for _, pid := range kids {
			if pid == victim {
				return false
			}
		}
		return get("site-00.test", "/")
	}) {
		return out, fmt.Errorf("child recovery failed")
	}
	out.CrashRecoveryMS = time.Since(start).Milliseconds()
	pids = children(cmd.Process.Pid)
	start = time.Now()
	signal := syscall.SIGTERM
	if killParent {
		signal = syscall.SIGKILL
	}
	if err = cmd.Process.Signal(signal); err != nil {
		return
	}
	select {
	case e := <-done:
		reaped = true
		if !killParent && e != nil {
			return out, e
		}
	case <-time.After(10 * time.Second):
		return out, fmt.Errorf("parent failed to stop")
	}
	if !waitUntil(5*time.Second, func() bool {
		for _, pid := range pids {
			if alive(pid) {
				return false
			}
		}
		return true
	}) {
		return out, fmt.Errorf("child survived parent exit")
	}
	out.ShutdownMS = time.Since(start).Milliseconds()
	if out.Failures != 0 {
		return out, fmt.Errorf("ordinary requests failed: %d", out.Failures)
	}
	return out, nil
}
