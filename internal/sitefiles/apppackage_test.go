package sitefiles

import (
	"testing"
)

func TestDecodeAppManifestStrict(t *testing.T) {
	valid := `{"name":"demo","version":"1.0.0","protocol":2,"executable":{"linux-amd64":"demo"}}`
	if manifest, err := DecodeAppManifest([]byte(valid)); err != nil || manifest.Name != "demo" {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	for _, tc := range []struct {
		name string
		raw  string
	}{
		{"duplicate", `{"name":"demo","version":"1.0.0","protocol":2,"executable":{"linux-amd64":"demo"},"protocol":2}`},
		{"unknown", `{"name":"demo","version":"1.0.0","protocol":2,"executable":{"linux-amd64":"demo"},"url":"x"}`},
		{"uppercase", `{"NAME":"demo","version":"1.0.0","protocol":2,"executable":{"linux-amd64":"demo"}}`},
		{"trailing", valid + "x"},
		{"missing key", `{"name":"demo","protocol":2,"executable":{"linux-amd64":"demo"}}`},
		{"not object", `[]`},
		{"empty", ``},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := DecodeAppManifest([]byte(tc.raw)); err == nil {
				t.Fatal("malformed manifest accepted")
			}
		})
	}
}

func TestValidateAppManifestRules(t *testing.T) {
	manifest := AppManifest{Name: "demo", Version: "1.0.0", Protocol: 2, Executable: map[string]string{"linux-amd64": "demo"}}
	if _, err := ValidateAppManifest(manifest, "demo", 2); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	base := manifest
	if _, err := ValidateAppManifest(base, "other", 2); err == nil {
		t.Fatal("name mismatch accepted")
	}
	if _, err := ValidateAppManifest(base, "demo", 1); err == nil {
		t.Fatal("protocol mismatch accepted")
	}
	empty := base
	empty.Version = ""
	if _, err := ValidateAppManifest(empty, "demo", 2); err == nil {
		t.Fatal("empty version accepted")
	}
	reserved := AppManifest{Name: "demo", Version: "1", Protocol: 2, Executable: map[string]string{"linux-amd64": "app.json"}}
	if _, err := ValidateAppManifest(reserved, "demo", 2); err == nil {
		t.Fatal("reserved app.json executable accepted")
	}
	data := AppManifest{Name: "demo", Version: "1", Protocol: 2, Executable: map[string]string{"linux-amd64": "data"}}
	if _, err := ValidateAppManifest(data, "demo", 2); err == nil {
		t.Fatal("runtime data executable accepted")
	}
	pathFile := AppManifest{Name: "demo", Version: "1", Protocol: 2, Executable: map[string]string{"linux-amd64": "../demo"}}
	if _, err := ValidateAppManifest(pathFile, "demo", 2); err == nil {
		t.Fatal("path executable accepted")
	}
}

func TestAppPackagePathPolicy(t *testing.T) {
	included := []string{"flea/apps/demo/app.json", "flea/apps/demo/demo", "flea/apps/demo/demo.exe"}
	for _, name := range included {
		if !AppPackageFile(name) {
			t.Fatalf("%s should be a package file", name)
		}
	}
	excluded := []string{"flea/apps/demo/data", "flea/apps/demo/.hidden", "flea/apps/demo/nested/demo", "flea/apps/Demo/app.json", "flea/apps/demo", "site/index.html"}
	for _, name := range excluded {
		if AppPackageFile(name) {
			t.Fatalf("%s should not be a package file", name)
		}
	}
	if !AppPackageDir("flea/apps") || !AppPackageDir("flea/apps/demo") {
		t.Fatal("package directories should be admitted")
	}
	if AppPackageDir("flea/apps/demo/nested") {
		t.Fatal("nested package directory admitted")
	}
	if !AppExecutablePath("flea/apps/demo/demo") || AppExecutablePath("flea/apps/demo/app.json") {
		t.Fatal("executable path classification wrong")
	}
}
