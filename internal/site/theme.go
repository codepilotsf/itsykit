package site

import (
	"flea/internal/snips"
	"flea/internal/theme"
	"fmt"
	"os"
)

// LoadWithState overlays separately managed live theme CSS on a code snapshot.
func LoadWithState(root string, content *snips.Store, css *string) (*Project, error) {
	return LoadWithMedia(root, content, css, "")
}
func LoadWithMedia(root string, content *snips.Store, css *string, mediaRoot string) (*Project, error) {
	p, err := loadProject(root, content, mediaRoot, false)
	if err != nil {
		return nil, err
	}
	if css != nil {
		if len(*css) > theme.MaxBytes {
			p.closeAssetFiles()
			return nil, fmt.Errorf("theme too large")
		}
		p.Assets["/css/theme.css"] = "site/css/theme.css"
		p.AssetData["/css/theme.css"] = []byte(*css)
	}
	return p, nil
}
func (p *Project) ThemeCSS() *string {
	b, ok := p.AssetData["/css/theme.css"]
	if !ok {
		return nil
	}
	s := string(b)
	return &s
}

// ReadManagedFile and WriteManagedFile reuse project confinement for fixed,
// caller-selected local/runtime paths. Never pass request-selected paths.
func ReadManagedFile(root, name string) ([]byte, error) {
	r, e := os.OpenRoot(root)
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return readFile(r, name)
}
func WriteManagedFile(root, name string, b []byte) error {
	r, e := os.OpenRoot(root)
	if e != nil {
		return e
	}
	defer r.Close()
	return writeSiteFile(r, name, b)
}
