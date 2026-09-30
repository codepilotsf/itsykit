package devaccess

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLANBoundary(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !LAN(r) {
			t.Error("validated request missing LAN authorization")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	h := WrapLAN(next, []string{"192.168.1.5", "10.0.0.5", "203.0.113.2"}, 3000)
	for _, tc := range []struct {
		host, peer string
		want       int
	}{
		{"192.168.1.5:3000", "192.168.1.9:4567", 204},
		{"10.0.0.5:3000", "10.0.0.9:4567", 204},
		{"localhost:3000", "127.0.0.1:4567", 204},
		{"127.0.0.1:3000", "127.0.0.1:4567", 204},
		{"evil.test:3000", "192.168.1.9:4567", 403},
		{"192.168.1.6:3000", "192.168.1.9:4567", 403},
		{"192.168.1.5:3001", "192.168.1.9:4567", 403},
		{"192.168.1.5", "192.168.1.9:4567", 403},
		{"192.168.1.5:3000", "203.0.113.9:4567", 403},
		{"203.0.113.2:3000", "192.168.1.9:4567", 403},
		{"192.168.1.5:3000", "invalid", 403},
	} {
		t.Run(tc.host+"/"+tc.peer, func(t *testing.T) {
			r := httptest.NewRequest("GET", "http://"+tc.host+"/", nil)
			r.RemoteAddr = tc.peer
			r.Header.Set("X-Forwarded-For", "192.168.1.9")
			r.Header.Set("X-Forwarded-Host", "192.168.1.5:3000")
			if LAN(r) {
				t.Fatal("headers authorized request")
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("got %d, want %d", w.Code, tc.want)
			}
		})
	}
}

func TestLANDefaultHTTPPort(t *testing.T) {
	h := WrapLAN(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }), []string{"192.168.1.5"}, 80)
	for _, host := range []string{"192.168.1.5", "192.168.1.5:80", "localhost", "127.0.0.1"} {
		r := httptest.NewRequest("GET", "http://"+host+"/", nil)
		r.RemoteAddr = "192.168.1.9:4567"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 204 {
			t.Fatalf("%s: %d", host, w.Code)
		}
	}
}
