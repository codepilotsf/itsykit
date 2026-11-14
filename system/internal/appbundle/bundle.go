// Package appbundle holds only the matching-platform Contact executable.
// The Contact executable must not import this package.
package appbundle

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"strings"
)

func Available() bool {
	_, err := fs.Stat(assets, "assets/"+runtime.GOOS+"-"+runtime.GOARCH+"/contact.gz")
	return err == nil
}

func Extract(ctx context.Context, destination string) error {
	base := "assets/" + runtime.GOOS + "-" + runtime.GOARCH + "/contact"
	sum, err := fs.ReadFile(assets, base+".sha256")
	if err != nil {
		return fmt.Errorf("Contact is not bundled. Rebuild with scripts/build.sh or install a complete release.")
	}
	expected, err := hex.DecodeString(strings.TrimSpace(string(sum)))
	if err != nil || len(expected) != 32 {
		return fmt.Errorf("Contact checksum is invalid")
	}
	source, err := assets.Open(base + ".gz")
	if err != nil {
		return fmt.Errorf("Contact executable is missing")
	}
	defer source.Close()
	z, err := gzip.NewReader(source)
	if err != nil {
		return fmt.Errorf("Contact bundle is damaged")
	}
	defer z.Close()
	f, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0500)
	if err != nil {
		return err
	}
	keep := false
	defer func() {
		f.Close()
		if !keep {
			os.Remove(destination)
		}
	}()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(&contextReader{ctx, z}, 20<<20+1))
	if err != nil || n > 20<<20 || hex.EncodeToString(h.Sum(nil)) != hex.EncodeToString(expected) {
		return fmt.Errorf("Contact bundle failed its integrity check")
	}
	if err := f.Close(); err != nil {
		return err
	}
	keep = true
	return nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
