package site

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"flea/internal/sitefiles"
	"flea/internal/snips"
)

// LoadEditing extracts new regions only for the local dev server.
func LoadEditing(ctx context.Context, root string) (*Project, error) {
	unlock := LockContent()
	defer unlock()
	return loadEditing(ctx, root)
}

// Caller holds the in-process content mutex through loading and publication.
func loadEditing(ctx context.Context, root string) (*Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	updated := false
	if err := extractSnips(root, &updated); err != nil {
		return nil, err
	}
	p, err := LoadDev(ctx, root)
	if err == nil {
		p.SnipsUpdated = updated
	}
	return p, err
}

// ExtractSnips is invoked by dev only, never by status or deployment packaging.
func ExtractSnips(directory string) error {
	unlock := LockContent()
	defer unlock()
	return extractSnips(directory, nil)
}

func extractSnips(directory string, updated *bool) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	// Retire only the empty file created by older dev versions. It no longer
	// coordinates writers; stop old dev processes before running this version.
	if private, e := pinnedDir(root, "_"); e == nil {
		if info, e := private.Lstat(".snips.lock"); e == nil && info.Mode().IsRegular() && info.Size() == 0 {
			_ = private.Remove(".snips.lock")
		}
		private.Close()
	}
	if data, e := readFile(root, transactionFile); e == nil {
		var tx snipTransaction
		if e = json.Unmarshal(data, &tx); e != nil {
			return e
		}
		if e = finishSnipTransaction(root, tx); e != nil {
			return e
		}
	} else if !os.IsNotExist(e) {
		return e
	}
	store, err := readSnipStore(root)
	if err != nil {
		return err
	}
	index, err := store.Index()
	if err != nil {
		return err
	}
	originalStore, _ := store.Bytes()
	references := map[string]string{}
	labels := map[string]string{}
	addReference := func(id, scope string) {
		if old, ok := references[id]; ok && old != scope {
			scope = ""
		}
		references[id] = scope
	}
	used := map[string]struct{}{}
	for id := range index {
		used[id] = struct{}{}
	}
	for id := range store.Archive {
		used[id] = struct{}{}
	}
	legacy := map[string]bool{}
	for _, directory := range []string{"_/snips", "_/data"} {
		entries, e := fs.ReadDir(root.FS(), directory)
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				name := strings.TrimSuffix(entry.Name(), ".json")
				if _, rich := used[name]; !rich {
					legacy[name] = true
				}
				used[name] = struct{}{}
			}
		}
	}
	tx := snipTransaction{Edits: map[string]snipEdit{}}
	previous, previousErr := readFile(root, "_/snips.json")
	if previousErr != nil && !os.IsNotExist(previousErr) {
		return previousErr
	}
	previousStore := snips.NewStore()
	if previousErr == nil {
		previousStore, err = snips.Decode(previous)
		if err != nil {
			return err
		}
	}
	previousSnapshot, _ := previousStore.Bytes()
	if !bytes.Equal(originalStore, previousSnapshot) {
		return fmt.Errorf("snips.json changed while preparing extraction; please save and retry")
	}
	tx.BeforeContent = previous
	err = fs.WalkDir(root.FS(), ".", func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !sitefiles.Included(name) || name == "_" {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(name, ".html") {
			return nil
		}
		data, e := readFile(root, name)
		if e != nil {
			return e
		}
		regions, ids, e := scanRegions(string(data), legacy)
		if e != nil {
			return fmt.Errorf("%s: %w", name, e)
		}
		for _, id := range ids {
			addReference(id, snipScope(name))
		}
		var out bytes.Buffer
		last := 0
		changed := false
		for _, r := range regions {
			if r.id != "" {
				labels[r.id] = r.label
				continue
			}
			id, e := snips.ReserveID(used)
			if e != nil {
				return e
			}
			content, e := snips.Sanitize(r.content)
			if e != nil {
				return e
			}
			store.Page(snipScope(name)).Snips[id] = snips.Entry{Label: r.label, Content: content}
			addReference(id, snipScope(name))
			labels[id] = r.label
			out.Write(data[last:r.start])
			out.WriteString(`{snip ` + id + `}`)
			last = r.end
			changed = true
		}
		if changed {
			out.Write(data[last:])
			tx.Edits[name] = snipEdit{Before: string(data), After: out.String()}
		}
		return nil
	})
	if err != nil {
		return err
	}
	index, err = store.Index()
	if err != nil {
		return err
	}
	for id, scope := range references {
		entry, ok := index[id]
		if !ok {
			entry, ok = store.Archive[id]
		}
		if !ok {
			return fmt.Errorf("missing snip %s in _/snips.json", id)
		}
		delete(store.Shared.Snips, id)
		for _, page := range store.Pages {
			delete(page.Snips, id)
		}
		delete(store.Archive, id)
		if label, ok := labels[id]; ok {
			entry.Label = label
		}
		store.Page(scope).Snips[id] = entry
	}
	// Keep removed content recoverable, but remove it from the active page index.
	// References include all templates, not just the page currently being rendered.
	store, err = snips.Merge(store, store, references)
	if err != nil {
		return err
	}
	tx.Content, err = store.Bytes()
	if err != nil {
		return err
	}
	if len(tx.Edits) == 0 && bytes.Equal(originalStore, tx.Content) {
		return nil
	}
	if _, err = snips.Decode(tx.Content); err != nil {
		return err
	}
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	if err = writeSiteFile(root, transactionFile, data); err != nil {
		return err
	}
	if err := finishSnipTransaction(root, tx); err != nil {
		return err
	}
	if updated != nil {
		*updated = true
	}
	return nil
}
