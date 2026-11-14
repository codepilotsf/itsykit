package appapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSessionDirective(t *testing.T) {
	good := strings.Repeat("a", 40)
	for _, tc := range []struct {
		name string
		raw  string
		want SessionDirective
		ok   bool
	}{
		{"omitted", "", SessionDirective{}, true},
		{"clear", "null", SessionDirective{Present: true, Clear: true}, true},
		{"set", `{"token":"` + good + `","max_age":60}`, SessionDirective{Present: true, Token: good, MaxAge: 60}, true},
		{"max age bound", `{"token":"` + good + `","max_age":2592000}`, SessionDirective{Present: true, Token: good, MaxAge: 2592000}, true},
		{"missing age", `{"token":"` + good + `"}`, SessionDirective{}, false},
		{"extra key", `{"token":"` + good + `","max_age":60,"extra":1}`, SessionDirective{}, false},
		{"short token", `{"token":"abc","max_age":60}`, SessionDirective{}, false},
		{"long token", `{"token":"` + strings.Repeat("a", 513) + `","max_age":60}`, SessionDirective{}, false},
		{"token space", `{"token":"aaaa aaaa aaaa aaaa","max_age":60}`, SessionDirective{}, false},
		{"token semicolon", `{"token":"aaaa;aaaaaaaaaaaaaaa","max_age":60}`, SessionDirective{}, false},
		{"zero age", `{"token":"` + good + `","max_age":0}`, SessionDirective{}, false},
		{"negative age", `{"token":"` + good + `","max_age":-1}`, SessionDirective{}, false},
		{"over age", `{"token":"` + good + `","max_age":2592001}`, SessionDirective{}, false},
		{"fractional age", `{"token":"` + good + `","max_age":1.5}`, SessionDirective{}, false},
		{"string age", `{"token":"` + good + `","max_age":"60"}`, SessionDirective{}, false},
		{"array", `[]`, SessionDirective{}, false},
		{"string", `"token"`, SessionDirective{}, false},
		{"uppercase token", `{"TOKEN":"` + good + `","max_age":60}`, SessionDirective{}, false},
		{"uppercase age", `{"token":"` + good + `","MAX_AGE":60}`, SessionDirective{}, false},
		{"trailing", `{"token":"` + good + `","max_age":60}x`, SessionDirective{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := Envelope{Session: []byte(tc.raw)}
			if tc.raw == "" {
				e.Session = nil
			}
			got, err := e.SessionDirective()
			if (err == nil) != tc.ok || got != tc.want {
				t.Fatalf("got %+v err=%v", got, err)
			}
		})
	}
}

func TestSessionTransport(t *testing.T) {
	const good = "session-token-0123456789abcdef"
	var seen string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "" {
			t.Error("browser cookie leaked to the app")
		}
		if r.URL.Path == "/_flea/handshake" {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"name":"contact","version":"1","protocol":2}`)
			return
		}
		seen = r.Header.Get(SessionHeader)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.RawQuery {
		case "clear":
			io.WriteString(w, `{"data":{},"session":null}`)
		case "set":
			io.WriteString(w, `{"data":{},"session":{"token":"`+good+`","max_age":60}}`)
		case "bad":
			io.WriteString(w, `{"data":{},"session":{"token":"short","max_age":60}}`)
		case "dup":
			io.WriteString(w, `{"data":{},"session":{"token":"`+good+`","max_age":60},"session":null}`)
		case "on422":
			w.WriteHeader(422)
			io.WriteString(w, `{"invalid":{"x":"y"},"session":{"token":"`+good+`","max_age":60}}`)
		default:
			io.WriteString(w, `{"data":{}}`)
		}
	})
	m := Mount{"contact", "/contact"}
	if err := c.Handshake(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.RequestWithSession(context.Background(), m, "GET", m.Path, "", nil, good); err != nil || seen != good {
		t.Fatalf("token not forwarded: %q %v", seen, err)
	}
	if _, _, err := c.RequestWithSession(context.Background(), m, "GET", m.Path, "", nil, "short"); err != nil || seen != "" {
		t.Fatalf("malformed token forwarded: %q %v", seen, err)
	}
	if _, _, err := c.Request(context.Background(), m, "GET", m.Path, "", nil); err != nil || seen != "" {
		t.Fatalf("token forwarded without a cookie: %q %v", seen, err)
	}
	e, _, err := c.RequestWithSession(context.Background(), m, "GET", m.Path, "set", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	d, err := e.SessionDirective()
	if err != nil || !d.Present || d.Clear || d.Token != good || d.MaxAge != 60 {
		t.Fatalf("set directive %+v %v", d, err)
	}
	e, _, err = c.RequestWithSession(context.Background(), m, "GET", m.Path, "clear", nil, good)
	if err != nil {
		t.Fatal(err)
	}
	if d, _ = e.SessionDirective(); !d.Present || !d.Clear {
		t.Fatalf("clear directive %+v", d)
	}
	for _, query := range []string{"bad", "dup"} {
		if _, _, err := c.RequestWithSession(context.Background(), m, "GET", m.Path, query, nil, ""); err != ErrUnavailable {
			t.Fatalf("%s accepted: %v", query, err)
		}
	}
	if _, _, err := c.RequestWithSession(context.Background(), m, "POST", m.Path, "on422", map[string]any{}, good); err != ErrUnavailable {
		t.Fatalf("session on 422 accepted: %v", err)
	}
	if _, ok := e.Values()["session"]; ok {
		t.Fatal("session leaked into template values")
	}
}
