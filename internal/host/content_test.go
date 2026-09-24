package host

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAliasesRequireOneContentSnapshot(t *testing.T) {
	file := testConfig(t, "127.0.0.1:8080")
	config, _, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	config.Sites["www.example.test"] = config.Sites["example.test"]
	content := filepath.Join(filepath.Dir(file), "content.json")
	if err := os.WriteFile(content, []byte(`{"pages":{},"shared":{"fleas":{}}}`), 0644); err != nil {
		t.Fatal(err)
	}
	for _, consistent := range []bool{false, true} {
		config.Content = map[string]string{"example.test": content}
		if consistent {
			config.Content["www.example.test"] = content
		}
		data, err := json.Marshal(config)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, data, 0644); err != nil {
			t.Fatal(err)
		}
		_, handler, err := Load(file, nil)
		if !consistent {
			if err == nil || !strings.Contains(err.Error(), "same content snapshot") {
				t.Fatalf("ambiguous alias: %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, hostname := range []string{"example.test", "WWW.EXAMPLE.TEST.:8080"} {
			r := httptest.NewRequest("GET", "/", nil)
			r.Host = hostname
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != 200 || !strings.Contains(w.Body.String(), "Host test") {
				t.Fatal(w.Code, w.Body.String())
			}
		}
	}
}
