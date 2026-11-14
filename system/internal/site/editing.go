package site

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"itsy/internal/snips"
)

// EditingVersion binds drafts to both content and developer-owned templates.
func (p *Project) EditingVersion() string {
	data, _ := json.Marshal(struct {
		Content *snips.Store
		Files   map[string]string
		App     map[string]any
		Data    map[string]map[string]any
		Root    string
		CSS     []byte
	}{p.Content, p.files, p.App, p.Data, p.Root, p.AssetData["/css/theme.css"]})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
