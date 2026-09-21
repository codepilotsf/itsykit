package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"nanoweb/internal/site"
	"nanoweb/templates"
)

const usage = "Usage: nanoweb new <site-name> | dev [-port 3000] | status\n"

// Run returns 0 on success, 1 on project/runtime failure, 2 on usage errors.
func Run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(errOut, usage)
		return 2
	}
	fail := func(err error) int { fmt.Fprintln(errOut, "nanoweb:", err); return 1 }
	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return 0
	case "new":
		if len(args) != 2 || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`).MatchString(args[1]) {
			fmt.Fprint(errOut, usage)
			return 2
		}
		if err := Create(args[1]); err != nil {
			return fail(err)
		}
		fmt.Fprintf(out, "\nCreated new project: %s\n\nNext steps:\n---------------\ncd %s\nnanoweb dev\n\n", args[1], args[1])
		return 0
	case "status":
		if len(args) != 1 {
			fmt.Fprint(errOut, usage)
			return 2
		}
		p, err := site.Load(".")
		if err != nil {
			return fail(err)
		}
		fmt.Fprintf(out, "Valid local project: %s\n%d pages, %d assets, %d snips\n", p.Site["name"], len(p.Pages), len(p.Assets), len(p.Snips))
		for _, route := range p.Routes() {
			fmt.Fprintln(out, " ", route)
		}
		return 0
	case "dev":
		flags := flag.NewFlagSet("dev", flag.ContinueOnError)
		flags.SetOutput(errOut)
		port := flags.Int("port", 3000, "local HTTP port (0 selects an available port)")
		if err := flags.Parse(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if flags.NArg() != 0 || *port < 0 || *port > 65535 {
			fmt.Fprint(errOut, usage)
			return 2
		}
		p, err := site.Load(".")
		if err != nil {
			return fail(err)
		}
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
		if err != nil {
			return fail(err)
		}
		handler := site.NewServer(p)
		server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		watchDone := make(chan struct{})
		go func() { defer close(watchDone); handler.Watch(watchCtx, p.Root, log.New(errOut, "", 0)) }()
		done := make(chan error, 1)
		go func() { done <- server.Serve(listener) }()
		fmt.Fprintf(out, "\nNanoweb running at http://%s\nWatching templates, assets, configuration and snips.\n\nPress Ctrl-C to stop.\n", listener.Addr())
		select {
		case <-ctx.Done():
			shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			if err := server.Shutdown(shutdownCtx); err != nil {
				server.Close()
			}
			<-done
		case err := <-done:
			cancel()
			<-watchDone
			if err != http.ErrServerClosed {
				return fail(err)
			}
			return 0
		}
		cancel()
		<-watchDone
		fmt.Fprintln(out, "Stopped.")
		return 0
	default:
		fmt.Fprint(errOut, usage)
		return 2
	}
}

// Create reserves a new directory first and never overwrites an existing target.
func Create(target string) error {
	if err := os.Mkdir(target, 0755); err != nil {
		return err
	}
	err := fs.WalkDir(templates.Files, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		body, err := templates.Files.ReadFile(name)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, err = f.Write(body)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		return nil
	})
	if err != nil {
		return err
	}
	_, err = site.Load(target)
	return err
}
