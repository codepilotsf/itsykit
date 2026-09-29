package setup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"flea/internal/backup"
	"flea/internal/media"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
)

// Fetch a coherent photo snapshot using the same private backup transport. The
// reviewed state digest prevents mixing photos from one revision with another.
func fetchPhotos(ctx context.Context, s DeploySettings, remote syncstate.Snapshot, ssh func(context.Context, string, bool, io.Writer) error, isRoot bool, temp string) (string, error) {
	if len(remote.Snips.ImageAssets()) == 0 && len(remote.Snips.VideoAssets()) == 0 {
		return "", nil
	}
	f, err := os.CreateTemp(temp, "photos-backup-")
	if err != nil {
		return "", err
	}
	defer f.Close()
	command := "/usr/local/bin/flea-server -backup -name " + shellQuote(s.Name) + " -branch " + shellQuote(s.Branch)
	if err = streamAdmin(ctx, command, ssh, isRoot, &boundedDownload{f, backup.MaxArchive + 4096}); err != nil {
		return "", err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(temp, "photos-")
	if err != nil {
		return "", err
	}
	manifest, err := backup.Extract(f, stage)
	if err != nil {
		return "", err
	}
	if manifest.Name != s.Name || manifest.Branch != s.Branch || manifest.StateVersion != remote.Version() {
		return "", fmt.Errorf("live photos changed; retry pull")
	}
	return filepath.Join(stage, "flea", "media"), nil
}

func installPhotos(root, source string, store *snips.Store) error {
	if len(store.VideoAssets()) > 0 {
		if err := site.IgnoreSessionFiles(root); err != nil {
			return err
		}
	}
	if len(store.ImageAssets()) == 0 && len(store.VideoAssets()) == 0 {
		return nil
	}
	dir := filepath.Join(root, "flea", "media")
	if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
		return err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("invalid media directory")
	}
	for _, photo := range store.ImageAssets() {
		b, e := site.ReadPhoto(dir, photo.File)
		if os.IsNotExist(e) && source != "" {
			b, e = site.ReadPhoto(source, photo.File)
			if e != nil {
				return e
			}
			if e = media.CheckBytes(photo, b); e != nil {
				return e
			}
			if e = site.WriteManagedFile(root, "flea/media/"+photo.File, b); e != nil {
				return e
			}
		}
		if e != nil {
			return e
		}
		if e = media.CheckBytes(photo, b); e != nil {
			return e
		}
	}
	return site.InstallVideos(source, dir, store)
}
