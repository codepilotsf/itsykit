package host

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func downloadFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "0.7.0")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	body := []byte("0123456789")
	if err := os.WriteFile(filepath.Join(dir, "flea-linux-amd64"), body, 0644); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf("%x  flea-linux-amd64\n", sha256.Sum256(body))
	if err := os.WriteFile(filepath.Join(dir, "SHA256SUMS"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCanonicalAndLegacyDownloadNamesShareVerifiedBytes(t *testing.T) {
	root := downloadFixture(t)
	version := filepath.Join(root, "0.7.0")
	body, err := os.ReadFile(filepath.Join(version, "flea-linux-amd64"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(version, "itsy-linux-amd64"), body, 0644); err != nil {
		t.Fatal(err)
	}
	manifest := fmt.Sprintf("%x  flea-linux-amd64\n%x  itsy-linux-amd64\n", sha256.Sum256(body), sha256.Sum256(body))
	if err := os.WriteFile(filepath.Join(version, "SHA256SUMS"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	d, err := loadDownloads(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"itsy-linux-amd64", "flea-linux-amd64"} {
		w := httptest.NewRecorder()
		d.ServeHTTP(w, httptest.NewRequest("GET", "http://example.test/releases/0.7.0/"+name, nil))
		if w.Code != 200 || w.Body.String() != string(body) {
			t.Fatal(name, w.Code)
		}
	}
}

func TestDownloadsRangeAndConfinement(t *testing.T) {
	d, err := loadDownloads(downloadFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path, rangeValue string
		status                   int
		body                     string
	}{
		{"GET", "/releases/0.7.0/flea-linux-amd64", "", 200, "0123456789"},
		{"HEAD", "/releases/0.7.0/flea-linux-amd64", "", 200, ""},
		{"GET", "/releases/0.7.0/flea-linux-amd64", "bytes=2-4", 206, "234"},
		{"POST", "/releases/0.7.0/flea-linux-amd64", "", 405, ""},
		{"GET", "/releases/0.7.0/", "", 404, "404 page not found\n"},
		{"GET", "/releases/0.7.0/../SHA256SUMS", "", 404, "404 page not found\n"},
		{"GET", "/releases/.env", "", 404, "404 page not found\n"},
		{"GET", "/releases/0.7.0/flea-linux-amd64/extra", "", 404, "404 page not found\n"},
	} {
		r := httptest.NewRequest(tc.method, tc.path, nil)
		r.Header.Set("Range", tc.rangeValue)
		w := httptest.NewRecorder()
		d.ServeHTTP(w, r)
		if w.Code != tc.status || w.Body.String() != tc.body {
			t.Fatalf("%+v: %d %q", tc, w.Code, w.Body.String())
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := httptest.NewRequest("GET", "/releases/0.7.0/flea-linux-amd64", nil)
			r.Header.Set("Range", "bytes=4-6")
			w := httptest.NewRecorder()
			d.ServeHTTP(w, r)
			if w.Body.String() != "456" {
				t.Error("concurrent range mismatch")
			}
		}()
	}
	wg.Wait()
}

func TestDownloadsRejectInvalidCatalogue(t *testing.T) {
	for _, kind := range []string{"checksum", "symlink", "hardlink", "manifest-link", "version-link", "unlisted", "writable", "oversize", "duplicate", "traversal"} {
		t.Run(kind, func(t *testing.T) {
			root := downloadFixture(t)
			binary := filepath.Join(root, "0.7.0", "flea-linux-amd64")
			manifest := filepath.Join(root, "0.7.0", "SHA256SUMS")
			var err error
			switch kind {
			case "checksum":
				err = os.WriteFile(binary, []byte("wrong"), 0644)
			case "symlink", "hardlink":
				outside := filepath.Join(t.TempDir(), "binary")
				if err = os.Rename(binary, outside); err == nil {
					if kind == "symlink" {
						err = os.Symlink(outside, binary)
					} else {
						err = os.Link(outside, binary)
					}
				}
			case "manifest-link":
				outside := filepath.Join(t.TempDir(), "manifest")
				if err = os.Rename(manifest, outside); err == nil {
					err = os.Symlink(outside, manifest)
				}
			case "version-link":
				err = os.Symlink(filepath.Join(root, "0.7.0"), filepath.Join(root, "0.8.0"))
			case "unlisted":
				err = os.WriteFile(filepath.Join(root, "0.7.0", "secret"), []byte("secret"), 0600)
			case "writable":
				err = os.Chmod(binary, 0666)
			case "oversize":
				err = os.Truncate(binary, maxDownloadBytes+1)
			case "duplicate":
				b, e := os.ReadFile(manifest)
				if e != nil {
					t.Fatal(e)
				}
				err = os.WriteFile(manifest, append(b, b...), 0644)
			case "traversal":
				err = os.WriteFile(manifest, []byte(strings.Repeat("0", 64)+"  ../secret\n"), 0644)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = loadDownloads(root); err == nil {
				t.Fatal("accepted invalid catalogue")
			}
		})
	}
}

func TestDownloadSnapshotMutation(t *testing.T) {
	root := downloadFixture(t)
	d, err := loadDownloads(root)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "0.7.0", "flea-linux-amd64")
	if err = os.Rename(file, file+".old"); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(file, []byte("replacement"), 0644); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	d.ServeHTTP(w, httptest.NewRequest("GET", "/releases/0.7.0/flea-linux-amd64", nil))
	if w.Body.String() != "0123456789" {
		t.Fatal("snapshot followed replaced path")
	}
	if err = os.Truncate(file+".old", 1); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	d.ServeHTTP(w, httptest.NewRequest("GET", "/releases/0.7.0/flea-linux-amd64", nil))
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}

func TestHostDownloadsIsolationAndLatest(t *testing.T) {
	file := testConfig(t, "127.0.0.1:8080")
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var c Config
	if err = json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	root := c.Sites["example.test"]
	if err = os.MkdirAll(filepath.Join(root, "site", "releases"), 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "site", "releases", "latest.txt"), []byte("0.7.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	c.Downloads = map[string]string{"example.test": downloadFixture(t)}
	save := func() {
		t.Helper()
		b, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(file, b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	save()
	_, h, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		host, path string
		status     int
		body       string
	}{
		{"example.test", "/releases/0.7.0/flea-linux-amd64", 200, "0123456789"},
		{"other.test", "/releases/0.7.0/flea-linux-amd64", 404, "404 page not found\n"},
		{"example.test", "/releases/latest.txt", 200, "0.7.0\n"},
	} {
		r := httptest.NewRequest("GET", tc.path, nil)
		r.Host = tc.host
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status || w.Body.String() != tc.body {
			t.Fatal(tc, w.Code, w.Body.String())
		}
	}
	c.Sites["alias.test"] = root
	save()
	if _, _, err = Load(file, nil); err == nil {
		t.Fatal("accepted mismatched alias store")
	}
	c.Downloads["alias.test"] = c.Downloads["example.test"]
	save()
	if _, _, err = Load(file, nil); err != nil {
		t.Fatal(err)
	}
	c.Downloads["unknown.test"] = c.Downloads["example.test"]
	save()
	if _, _, err = Load(file, nil); err == nil {
		t.Fatal("accepted unknown host")
	}
}
