package site

import (
	"bytes"
	"log"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRequestLogging(t *testing.T) {
	root := fixture(t)
	var output bytes.Buffer
	s := NewServer(mustLoad(t, root), log.New(&output, "", 0))
	for _, test := range []struct{ method, path string }{
		{"GET", "/"}, {"GET", "/about?token=secret"}, {"HEAD", "/about"},
		{"GET", "/missing"}, {"POST", "/about"}, {"GET", "/css/main.css"},
		{"GET", reloadPath}, {"HEAD", reloadPath}, {"GET", "/bad%0Aline"},
	} {
		request(s, test.method, test.path)
	}
	r := httptest.NewRequest("GET", "/css/main.css", nil)
	r.Header.Set("Range", "bytes=0-3")
	s.ServeHTTP(httptest.NewRecorder(), r)
	put(t, root, "index.html", "{bad syntax}")
	if err := s.Reload(root); err == nil {
		t.Fatal("expected validation error")
	}
	request(s, "GET", "/")
	want := "GET /\nGET /about\nHEAD /about\n404 GET /missing\n405 POST /about\nGET /css/main.css\n404 GET /bad%0Aline\n206 GET /css/main.css\n500 GET /\n"
	if output.String() != want {
		t.Fatalf("got:\n%swant:\n%s", output.String(), want)
	}
}

func TestConcurrentRequestLogs(t *testing.T) {
	var output bytes.Buffer
	s := NewServer(mustLoad(t, fixture(t)), log.New(&output, "", 0))
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); request(s, "GET", "/about") }()
	}
	wg.Wait()
	if strings.Count(output.String(), "GET /about\n") != 20 {
		t.Fatal(output.String())
	}
}
