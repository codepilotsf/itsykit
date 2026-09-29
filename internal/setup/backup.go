package setup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"flea/internal/backup"
	"flea/internal/site"
)

type boundedDownload struct {
	dst       io.Writer
	remaining int64
}

func (w *boundedDownload) Write(p []byte) (int, error) {
	if int64(len(p)) > w.remaining {
		return 0, fmt.Errorf("backup exceeds size limit")
	}
	n, e := w.dst.Write(p)
	w.remaining -= int64(n)
	return n, e
}

func Backup(ctx context.Context, root, file string, in io.Reader, out, errOut io.Writer) error {
	return backupSite(ctx, root, file, in, out, errOut, execute)
}
func backupSite(ctx context.Context, root, file string, in io.Reader, out, errOut io.Writer, run runner) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("choose a new backup filename: %w", err)
	}
	complete := false
	defer func() {
		f.Close()
		if !complete {
			os.Remove(file)
		}
	}()
	err = withRemoteAdmin(ctx, root, in, out, errOut, run, func(r remoteAdmin) error {
		cmd := "/usr/local/bin/flea-server -backup -name " + shellQuote(r.settings.Name) + " -branch " + shellQuote(r.settings.Branch)
		if e := streamAdmin(ctx, cmd, r.ssh, r.root, &boundedDownload{f, backup.MaxArchive + 4096}); e != nil {
			return e
		}
		if e := f.Sync(); e != nil {
			return e
		}
		if e := f.Close(); e != nil {
			return e
		}
		input, e := os.Open(file)
		if e != nil {
			return e
		}
		defer input.Close()
		stage, e := os.MkdirTemp("", "flea-verify-backup-")
		if e != nil {
			return e
		}
		defer os.RemoveAll(stage)
		manifest, e := backup.Extract(input, stage)
		if e != nil {
			return e
		}
		if manifest.Name != r.settings.Name || manifest.Branch != r.settings.Branch {
			return fmt.Errorf("backup identity mismatch")
		}
		_, e = site.Load(stage)
		return e
	})
	if err != nil {
		return err
	}
	complete = true
	fmt.Fprintf(out, "Backup saved: %s\n", file)
	return nil
}

func Restore(ctx context.Context, root, file string, in io.Reader, out, errOut io.Writer) error {
	return restoreSite(ctx, root, file, in, out, errOut, execute)
}
func restoreSite(ctx context.Context, root, file string, in io.Reader, out, errOut io.Writer, run runner) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	// Pin a private copy so the uploaded bytes cannot differ from reviewed bytes.
	pinned, err := os.CreateTemp("", "flea-restore-input-")
	if err != nil {
		return err
	}
	defer os.Remove(pinned.Name())
	defer pinned.Close()
	n, err := io.Copy(pinned, io.LimitReader(f, backup.MaxArchive+4097))
	if err != nil {
		return err
	}
	if n > backup.MaxArchive+4096 {
		return fmt.Errorf("backup exceeds size limit")
	}
	if _, err = pinned.Seek(0, io.SeekStart); err != nil {
		return err
	}
	stage, err := os.MkdirTemp("", "flea-review-restore-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	manifest, err := backup.Extract(pinned, stage)
	if err != nil {
		return err
	}
	if _, err = site.Load(stage); err != nil {
		return err
	}
	return withRemoteAdmin(ctx, root, in, out, errOut, run, func(r remoteAdmin) error {
		if manifest.Name != r.settings.Name || manifest.Branch != r.settings.Branch {
			return fmt.Errorf("backup belongs to another site or branch")
		}
		live, e := fetchState(ctx, r.settings, r.ssh, r.root)
		if e != nil {
			return e
		}
		if !live.Exists {
			return fmt.Errorf("deploy this branch before restoring its backup")
		}
		fmt.Fprintf(out, "Restore %s: replace live code, configuration, snips and theme with this backup.\n", r.settings.Domain)
		answer, e := ask(ctx, in, out, "Type RESTORE to continue", "")
		if e != nil {
			return e
		}
		if answer != "RESTORE" {
			return fmt.Errorf("cancelled; nothing changed")
		}
		current, e := currentGit(ctx, root, false)
		if e != nil {
			return e
		}
		if current.Branch != r.settings.Branch {
			return fmt.Errorf("branch changed; retry")
		}
		var path bytes.Buffer
		if e = r.ssh(ctx, "umask 077; mktemp /tmp/flea-restore.XXXXXXXXXX", false, &path); e != nil {
			return e
		}
		remote := strings.TrimSpace(path.String())
		if !regexp.MustCompile(`^/tmp/flea-restore\.[A-Za-z0-9]{10}$`).MatchString(remote) {
			return fmt.Errorf("unexpected restore staging path")
		}
		defer func() {
			cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			_ = r.ssh(cleanup, "rm -f -- "+shellQuote(remote), false, io.Discard)
		}()
		if _, e = pinned.Seek(0, io.SeekStart); e != nil {
			return e
		}
		if e = r.upload(ctx, remote, pinned); e != nil {
			return e
		}
		cmd := "/usr/local/bin/flea-server -restore " + shellQuote(remote) + " -name " + shellQuote(r.settings.Name) + " -branch " + shellQuote(r.settings.Branch) + " -expected-state " + shellQuote(live.Version())
		if !r.root {
			cmd = "sudo -- " + cmd
		}
		return r.ssh(ctx, cmd, !r.root, out)
	})
}
