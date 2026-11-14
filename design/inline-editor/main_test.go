package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrototypeBoundaries(t *testing.T) {
	for _, test := range []struct {
		method, path, host, origin, body string
		want                             int
	}{
		{"GET", "/", "127.0.0.1:3143", "", "", 200},
		{"GET", "/main.go", "127.0.0.1:3143", "", "", 404},
		{"GET", "/site/.itsy.config.json", "127.0.0.1:3143", "", "", 404},
		{"GET", "/", "example.com:3143", "", "", 403},
		{"POST", "/sanitize-check", "127.0.0.1:3143", "http://other.test", "<p>Hello</p>", 403},
		{"POST", "/sanitize-check", "127.0.0.1:3143", "http://127.0.0.1:3143", "<h1>Title</h1>", 200},
	} {
		r := httptest.NewRequest(test.method, "http://"+test.host+test.path, strings.NewReader(test.body))
		r.Header.Set("Origin", test.origin)
		w := httptest.NewRecorder()
		handler().ServeHTTP(w, r)
		if w.Code != test.want {
			t.Fatalf("%s %s: %d %s", test.method, test.path, w.Code, w.Body.String())
		}
	}
}
