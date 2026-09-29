// Package backup defines portable, bounded site backups. Credentials and runtime
// sessions are excluded by the release inclusion policy, not by caller choice.
package backup

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"flea/internal/release"
)

const magic = "FLEA-BACKUP-1\n"
const MaxArchive int64 = 270 << 20

type Manifest struct {
	StateVersion string `json:"state_version,omitempty"`
	Name         string `json:"name"`
	Branch       string `json:"branch"`
	Commit       string `json:"commit"`
	Size         int64  `json:"size"`
	SHA256       string `json:"sha256"`
}

func (m Manifest) validate() error {
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`).MatchString(m.Name) || m.Branch == "" || len(m.Branch) > 1024 || !regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`).MatchString(m.Commit) {
		return fmt.Errorf("invalid backup identity")
	}
	if m.Size <= 0 || m.Size > MaxArchive || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(m.SHA256) {
		return fmt.Errorf("invalid backup payload")
	}
	return nil
}

func Write(root string, m Manifest, out io.Writer) error {
	f, err := os.CreateTemp("", "flea-backup-payload-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	h := sha256.New()
	if err = release.Pack(root, io.MultiWriter(f, h)); err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	m.Size = info.Size()
	m.SHA256 = hex.EncodeToString(h.Sum(nil))
	if err = m.validate(); err != nil {
		return err
	}
	if _, err = io.WriteString(out, magic); err != nil {
		return err
	}
	if err = json.NewEncoder(out).Encode(m); err != nil {
		return err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	_, err = io.Copy(out, f)
	return err
}

// Extract requires a fresh private destination. On failure, discard it entirely.
func Extract(input io.Reader, dest string) (Manifest, error) {
	var m Manifest
	r := bufio.NewReaderSize(input, 4096)
	header := make([]byte, len(magic))
	_, err := io.ReadFull(r, header)
	if err != nil || string(header) != magic {
		return m, fmt.Errorf("not a Flea backup")
	}
	// ReadSlice bounds an untrusted manifest before allocating a JSON document.
	raw, err := r.ReadSlice('\n')
	if err != nil {
		return m, fmt.Errorf("invalid backup manifest")
	}
	if err = json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	if err = m.validate(); err != nil {
		return m, err
	}
	h := sha256.New()
	limited := &io.LimitedReader{R: r, N: m.Size}
	err = release.Extract(io.TeeReader(limited, h), dest)
	if err != nil {
		return m, err
	}
	if limited.N != 0 || hex.EncodeToString(h.Sum(nil)) != m.SHA256 {
		return m, fmt.Errorf("backup is incomplete or corrupted")
	}
	if _, err = r.ReadByte(); err != io.EOF {
		return m, fmt.Errorf("unexpected data after backup")
	}
	return m, nil
}
