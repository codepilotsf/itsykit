package contact

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFlows(t *testing.T) {
	for _, tc := range []struct {
		body          string
		fail          bool
		status, calls int
	}{
		{`{"name":"Alex","email":"a@example.test","message":"Hello"}`, false, 200, 1},
		{`{"name":"Alex","email":"a@example.test","message":"Hello"}`, true, 503, 1},
		{`{"name":"Alex","email":"bad","message":"Hello"}`, false, 422, 0},
		{`{"name":["Alex","Other"],"email":"a@example.test","message":"Hello"}`, false, 422, 0},
		{`{"name":"Alex","email":"a@example.test","message":"Hello","website":"spam"}`, false, 200, 0},
		{`{"name":"Alex","email":"a@example.test","message":"Hello","to":"attacker"}`, false, 422, 0},
		{`{"name":"Alex","name":"Other"}`, false, 400, 0},
	} {
		calls := 0
		h := Handler{Notify: func(ctx context.Context, s Submission) error {
			calls++
			if s.Name != "Alex" || s.Email != "a@example.test" || s.Message != "Hello" {
				t.Error(s)
			}
			if tc.fail {
				return errors.New("provider-secret")
			}
			return nil
		}}
		r := httptest.NewRequest("POST", "http://127.0.0.1/", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status || calls != tc.calls || strings.Contains(w.Body.String(), "provider-secret") || strings.Contains(w.Body.String(), "attacker") {
			t.Fatal(w.Code, w.Body.String(), calls)
		}
		if w.Code == 200 && !strings.Contains(w.Body.String(), `"redirect":"thank-you"`) {
			t.Fatal("success must use mount-relative thank-you", w.Body.String())
		}
	}
}
