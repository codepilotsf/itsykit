package setup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"flea/internal/release"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
)

type limitedOutput struct{ bytes.Buffer }

func (b *limitedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 8<<20 {
		return 0, fmt.Errorf("remote state exceeds size limit")
	}
	return b.Buffer.Write(p)
}
func destination(s DeploySettings) string {
	return s.Target + "\n" + s.Name + "\n" + s.Domain + "\n" + s.Branch
}
func baselinePath(ctx context.Context, root string, s DeploySettings) (string, error) {
	path, e := git(ctx, root, "rev-parse", "--path-format=absolute", "--git-path", "flea-sync")
	if e != nil {
		return "", e
	}
	path = strings.TrimSpace(path)
	h := sha256.Sum256([]byte(destination(s) + "\n" + root))
	return filepath.Join(path, hex.EncodeToString(h[:])+".json"), nil
}
func readBaseline(ctx context.Context, root string, s DeploySettings) (syncstate.Baseline, error) {
	b := syncstate.Baseline{Destination: destination(s)}
	p, e := baselinePath(ctx, root, s)
	if e != nil {
		return b, e
	}
	raw, e := site.ReadManagedFile(filepath.Dir(p), filepath.Base(p))
	if os.IsNotExist(e) {
		return b, nil
	}
	if e != nil {
		return b, e
	}
	if len(raw) > 8<<20 {
		return b, fmt.Errorf("sync baseline too large")
	}
	if e = json.Unmarshal(raw, &b); e != nil {
		return b, e
	}
	if b.Destination != destination(s) {
		return b, fmt.Errorf("sync destination changed; pull again")
	}
	return b, nil
}
func saveBaseline(ctx context.Context, root string, s DeploySettings, b syncstate.Baseline) error {
	p, e := baselinePath(ctx, root, s)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		return e
	}
	b.Destination = destination(s)
	raw, e := json.Marshal(b)
	if e != nil {
		return e
	}
	return site.WriteManagedFile(filepath.Dir(p), filepath.Base(p), raw)
}
func resolveSync(ctx context.Context, in io.Reader, out io.Writer) syncstate.Resolve {
	return func(c syncstate.Conflict) (bool, error) {
		if file, ok := in.(*os.File); ok {
			check := exec.CommandContext(ctx, "stty", "-g")
			check.Stdin = file
			check.Stdout = io.Discard
			check.Stderr = io.Discard
			if check.Run() != nil {
				return false, fmt.Errorf("conflicting %s; rerun in an interactive terminal", c.Key)
			}
		}
		// Quoting prevents terminal control sequences from remote content executing.
		fmt.Fprintf(out, "\nConflict: %s\n\nLocal: %q\nLive:  %q\n\n", c.Key, c.Local, c.Remote)
		for {
			v, e := ask(ctx, in, out, "Keep [l]ive, use [d]eveloper version, or [c]ancel", "")
			if e != nil {
				return false, fmt.Errorf("unresolved conflict; nothing published")
			}
			switch strings.ToLower(v) {
			case "l", "live":
				return false, nil
			case "d", "local", "developer":
				return true, nil
			case "c", "cancel":
				return false, fmt.Errorf("cancelled; nothing published")
			}
		}
	}
}

func fetchState(ctx context.Context, s DeploySettings, ssh func(context.Context, string, bool, io.Writer) error, isRoot bool) (syncstate.Snapshot, error) {
	var remote syncstate.Snapshot
	command := "/usr/local/bin/flea-server -export-state -name " + shellQuote(s.Name) + " -branch " + shellQuote(s.Branch)
	out, e := captureAdmin(ctx, command, ssh, isRoot)
	if e != nil {
		return remote, e
	}
	d := json.NewDecoder(bytes.NewReader(out))
	d.DisallowUnknownFields()
	if e := d.Decode(&remote); e != nil {
		return remote, e
	}
	if d.Decode(new(any)) != io.EOF || remote.Snips == nil {
		return remote, fmt.Errorf("invalid live state")
	}
	b, e := remote.Snips.Bytes()
	if e != nil {
		return remote, e
	}
	if _, e = snips.Decode(b); e != nil {
		return remote, e
	}
	return remote, nil
}

func captureAdmin(ctx context.Context, command string, ssh func(context.Context, string, bool, io.Writer) error, isRoot bool) ([]byte, error) {
	var out limitedOutput
	err := streamAdmin(ctx, command, ssh, isRoot, &out)
	return out.Bytes(), err
}
func streamAdmin(ctx context.Context, command string, ssh func(context.Context, string, bool, io.Writer) error, isRoot bool, out io.Writer) error {
	if !isRoot {
		// Keep password prompts on the terminal and JSON on a separate channel.
		// The SSH user owns this private temporary file, not a public endpoint.
		var path bytes.Buffer
		if e := ssh(ctx, "umask 077; mktemp /tmp/flea-state.XXXXXXXXXX", false, &path); e != nil {
			return e
		}
		stage := strings.TrimSpace(path.String())
		if !regexp.MustCompile(`^/tmp/flea-state\.[A-Za-z0-9]{10}$`).MatchString(stage) {
			return fmt.Errorf("unexpected state staging path")
		}
		defer func() {
			cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			_ = ssh(cleanup, "rm -f -- "+shellQuote(stage), false, io.Discard)
		}()
		if e := ssh(ctx, "sudo -- "+command+" > "+shellQuote(stage), true, os.Stderr); e != nil {
			return e
		}
		command = "cat -- " + shellQuote(stage)
	}
	if e := ssh(ctx, command, false, out); e != nil {
		return fmt.Errorf("could not read live state: %w", e)
	}
	return nil
}

func makePlan(base syncstate.Baseline, local *site.Project, remote syncstate.Snapshot, withSnips bool, resolve syncstate.Resolve) (*syncstate.Plan, error) {
	plan := &syncstate.Plan{Expected: remote.Version()}
	var e error
	if withSnips {
		if !remote.Exists {
			plan.Snips = local.Content
		} else {
			plan.Snips, e = syncstate.MergeSnips(base.Snips, local.Content, remote.Snips, resolve)
			if e != nil {
				return nil, e
			}
		}
	}
	return plan, nil
}

func Pull(ctx context.Context, root string, in io.Reader, out, errOut io.Writer) error {
	return pull(ctx, root, in, out, errOut, execute)
}
func pull(ctx context.Context, root string, in io.Reader, out, errOut io.Writer, run runner) error {
	root, e := filepath.Abs(root)
	if e != nil {
		return e
	}
	unlock, e := site.LockDevelopment(root)
	if e != nil {
		return e
	}
	defer unlock()
	g, e := currentGit(ctx, root, false)
	if e != nil {
		return e
	}
	c, e := readDeploymentConfig(root)
	if e != nil {
		return e
	}
	s, ok := c.Branches[g.Branch]
	if !ok {
		return fmt.Errorf("configure this branch with flea deploy -configure first")
	}
	s.Branch = g.Branch
	if e = validateMapping(g.Branch, s, c); e != nil {
		return e
	}
	// Pull never overwrites uncommitted content, and never commits on your behalf.
	paths := []string{"flea/snips.json", "flea/media/"}
	args := append([]string{"status", "--porcelain=v1", "--untracked-files=all", "--"}, paths...)
	status, e := git(ctx, root, args...)
	if e != nil {
		return e
	}
	if status != "" {
		return fmt.Errorf("commit your local snip changes before pulling")
	}
	p, e := site.LoadSyncSource(root)
	if e != nil {
		return e
	}
	if e = p.CheckExtractedSnips(); e != nil {
		return e
	}
	base, e := readBaseline(ctx, root, s)
	if e != nil {
		return e
	}
	target, e := targetName(s.Target)
	if e != nil {
		return e
	}
	temp, e := os.MkdirTemp("", "flea-pull-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(temp)
	options := connectionOptions(temp)
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		args := append(append([]string{}, options...), "-O", "exit", "--", target)
		_ = run(cleanup, command{name: "ssh", args: args, out: io.Discard, errOut: io.Discard})
	}()
	ssh := func(ctx context.Context, text string, tty bool, w io.Writer) error {
		args := append([]string{}, options...)
		if tty {
			args = append(args, "-tt")
		} else {
			args = append(args, "-T")
		}
		args = append(args, "--", target, text)
		return run(ctx, command{name: "ssh", args: args, in: in, out: w, errOut: errOut})
	}
	var probeOut bytes.Buffer
	if e = ssh(ctx, "/usr/local/bin/flea-server -protocol", false, &probeOut); e != nil {
		return e
	}
	if strings.TrimSpace(probeOut.String()) != release.Protocol {
		return fmt.Errorf("update the server before pulling")
	}
	isRoot := strings.HasPrefix(target, "root@")
	remote, e := fetchState(ctx, s, ssh, isRoot)
	if e != nil {
		return e
	}
	if !remote.Exists {
		return fmt.Errorf("this branch has not been deployed yet")
	}
	photoSource, e := fetchPhotos(ctx, s, remote, ssh, isRoot, temp)
	if e != nil {
		return e
	}
	nextContent := p.Content
	comparison := base.Snips
	if comparison == nil {
		comparison = snips.NewStore()
	}
	merged, e := syncstate.MergeSnips(comparison, p.Content, remote.Snips, resolveSync(ctx, in, out))
	if e != nil {
		return e
	}
	refs, e := p.SnipReferences()
	if e != nil {
		return e
	}
	nextContent, e = snips.Merge(p.Content, merged, refs)
	if e != nil {
		return e
	}
	base.Snips = remote.Snips
	// Check again after network/prompt time. Never publish a baseline before files.
	latest, e := currentGit(ctx, root, false)
	if e != nil {
		return e
	}
	status, e = git(ctx, root, args...)
	if e != nil {
		return e
	}
	if latest.Branch != g.Branch || latest.Commit != g.Commit || status != "" {
		return fmt.Errorf("local files or branch changed during pull; retry")
	}
	if e = installPhotos(root, photoSource, nextContent); e != nil {
		return e
	}
	if _, e = site.LoadWithSnips(root, nextContent); e != nil {
		return e
	}
	if e = site.SaveContent(root, nextContent); e != nil {
		return e
	}
	if e = saveBaseline(ctx, root, s, base); e != nil {
		return fmt.Errorf("files pulled, but baseline could not be saved: %w", e)
	}
	fmt.Fprintln(out, "Pulled live snips and media. Review and commit the local changes.")
	return nil
}
