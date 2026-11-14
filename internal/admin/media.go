package admin

import (
	"fmt"
	"itsy/internal/media"
	"itsy/internal/site"
	"itsy/internal/sitefiles"
	"itsy/internal/snips"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

func (m Manager) prepareMedia(source, content string, store *snips.Store) error {
	if len(store.ImageAssets()) == 0 && len(store.VideoAssets()) == 0 {
		return nil
	}
	root := site.ManagedMediaRoot(content)
	if err := realDirs(filepath.Dir(root), []string{"media"}); err != nil {
		return err
	}
	if m.ConfigPath == "/etc/flea/server.json" {
		account, err := user.Lookup("flea")
		if err != nil {
			return err
		}
		uid, err := strconv.Atoi(account.Uid)
		if err != nil {
			return err
		}
		gid, err := strconv.Atoi(account.Gid)
		if err != nil {
			return err
		}
		if err = os.Chown(root, uid, gid); err != nil {
			return err
		}
		if err = os.Chmod(root, 0750); err != nil {
			return err
		}
	}
	total := 0
	seen := make(map[string]bool)
	for _, photo := range store.ImageAssets() {
		data, err := site.ReadPhoto(root, photo.File)
		if os.IsNotExist(err) {
			data, err = site.ReadPhoto(sitefiles.Path(source, filepath.Join("itsy", "media")), photo.File)
			if err != nil {
				return fmt.Errorf("missing photo in deployment: %w", err)
			}
			if err = media.CheckBytes(photo, data); err != nil {
				return err
			}
			if err = site.WriteManagedFile(root, photo.File, data); err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
		if err = media.CheckBytes(photo, data); err != nil {
			return err
		}
		if !seen[photo.File] {
			total += len(data)
			seen[photo.File] = true
		}
		if total > media.MaxTotal {
			return fmt.Errorf("photo library exceeds 64 MiB")
		}
	}
	if err := site.InstallVideos(sitefiles.Path(source, filepath.Join("itsy", "media")), root, store); err != nil {
		return err
	}
	return nil
}
