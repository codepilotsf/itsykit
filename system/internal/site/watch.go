package site

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"itsy/internal/sitefiles"
)

const watchDebounce = 100 * time.Millisecond

// Watch uses directory notifications, including atomic file replacements. A
// fingerprint is checked only after events (and once after watch registration).
func (s *Server) Watch(ctx context.Context, root string, logger *log.Logger) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("Could not watch the project folder: %w", err)
	}
	defer watcher.Close()
	if err := syncWatches(watcher, root); err != nil {
		return err
	}
	timer := time.NewTimer(watchDebounce)
	defer timer.Stop()
	s.mu.RLock()
	last := s.project.signature
	s.mu.RUnlock()
	previous := ""
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("Project watching stopped. Restart itsy dev.")
			}
			rel, err := filepath.Rel(root, event.Name)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue
			}
			if rel != "." && !sitefiles.Included(filepath.ToSlash(rel)) {
				continue
			}
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				// Forget moved watches, even when the old path has already been recreated.
				for _, name := range watcher.WatchList() {
					if name == event.Name || strings.HasPrefix(name, event.Name+string(filepath.Separator)) {
						_ = watcher.Remove(name)
					}
				}
			}
			timer.Reset(watchDebounce)
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("Project watching stopped. Restart itsy dev.")
			}
			if !errors.Is(err, fsnotify.ErrEventOverflow) {
				return fmt.Errorf("Could not watch the project folder: %w", err)
			}
			// The OS lost events. Reconcile watches and content once, not periodically.
			// A lost rename can leave a watch attached to an old directory inode.
			for _, name := range watcher.WatchList() {
				_ = watcher.Remove(name)
			}
			if err := syncWatches(watcher, root); err != nil {
				return err
			}
			timer.Reset(watchDebounce)
		case <-timer.C:
			if err := syncWatches(watcher, root); err != nil {
				return err
			}
			next := fingerprint(root)
			if next == last {
				continue
			}
			last = next
			err := s.reload(ctx, root, true)
			if ctx.Err() != nil {
				return nil
			}
			if err == nil {
				s.mu.RLock()
				project := s.project
				last = project.signature
				s.mu.RUnlock()
				if logger != nil {
					for _, warning := range project.Warnings {
						logger.Printf("Warning: %s", warning)
					}
				}
			}
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != previous && logger != nil {
				if message == "" {
					logger.Print("Site valid again; changes loaded")
				} else {
					logger.Printf("Site invalid: %q", message)
				}
			}
			previous = message
		}
	}
}

// Register parents before walking children, so changes during discovery either
// appear in the scan or generate an event. Never descend into excluded paths or
// symlinks. The normal confined loader remains the authority on valid content.
func syncWatches(w *fsnotify.Watcher, root string) error {
	watched := map[string]bool{}
	for _, name := range w.WatchList() {
		watched[name] = true
	}
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			if name != root && os.IsNotExist(err) {
				return nil
			}
			return err
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		if !sitefiles.Included(filepath.ToSlash(rel)) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.IsDir() {
			if name == root {
				return fmt.Errorf("project folder is no longer a directory")
			}
			return nil
		}
		seen[name] = true
		if !watched[name] {
			if err := w.Add(name); err != nil {
				if name != root && os.IsNotExist(err) {
					return fs.SkipDir
				}
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Could not watch the project folder. Check its permissions and restart itsy dev: %w", err)
	}
	for name := range watched {
		if !seen[name] {
			_ = w.Remove(name)
		}
	}
	return nil
}
