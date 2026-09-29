package site

import (
	"context"
	"flea/internal/snips"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOwnerDevRevisionTracksExtraction(t *testing.T) {
	root := fixture(t)
	p, err := LoadEditing(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(p, nil)
	render := func(s *Server) (string, string) {
		w := httptest.NewRecorder()
		s.ServeOwner(w, httptest.NewRequest("GET", "/", nil), func(body []byte, _ *snips.Store, _ string) ([]byte, bool) { return body, true })
		return w.Header().Get("X-Flea-Dev-Revision"), w.Body.String()
	}
	before, _ := render(server)
	if before == "" {
		t.Fatal("missing dev revision")
	}
	put(t, root, "index.html", `<p snip>New region</p>`)
	if err := server.reload(context.Background(), root, true); err != nil {
		t.Fatal(err)
	}
	after, body := render(server)
	if after == before || !strings.Contains(body, "snip-start:") {
		t.Fatal("new region/revision missing", body)
	}
	if revision, _ := render(NewProductionServer(p, nil)); revision != "" {
		t.Fatal("production reload enabled")
	}
}
