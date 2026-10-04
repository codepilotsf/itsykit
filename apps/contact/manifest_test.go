package contact

import (
	"encoding/json"
	"flea/internal/appapi"
	"os"
	"testing"
)

func TestManifestMatchesHandshake(t *testing.T) {
	data, err := os.ReadFile("app.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest appapi.Identity
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != "contact" || manifest.Version != Version || manifest.Protocol != appapi.Version {
		t.Fatal("manifest and handshake disagree")
	}
}
