// Package contact is the included, storage-free Contact mini app.
// Delivery is a narrow capability supplied by Flea, not provider access.
package contact

import (
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"flea/internal/appapi"
)

const Version = "0.1.0"

type Submission struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

func (s Submission) Data() map[string]any {
	return map[string]any{"name": s.Name, "email": s.Email, "message": s.Message}
}

// Redisplay excludes unrecognized fields, repeated values and the spam trap.
// Bound each value independently even when the app is unavailable.
func Redisplay(fields map[string]any) map[string]any {
	data := map[string]any{}
	for name, max := range map[string]int{"name": 100, "email": 254, "message": 5000} {
		value, ok := fields[name].(string)
		if ok && utf8.ValidString(value) && utf8.RuneCountInString(value) <= max {
			data[name] = value
		}
	}
	return data
}

func Validate(fields map[string]any) (Submission, map[string]string, bool) {
	invalid := map[string]string{}
	values := map[string]string{}
	for name, value := range fields {
		if name != "name" && name != "email" && name != "message" && name != "website" {
			invalid[name] = "This field is not supported."
			continue
		}
		text, ok := value.(string)
		if !ok || !utf8.ValidString(text) {
			invalid[name] = "Use one text value for this field."
			continue
		}
		values[name] = strings.TrimSpace(text)
	}
	s := Submission{Name: values["name"], Email: values["email"], Message: values["message"]}
	if s.Name == "" || utf8.RuneCountInString(s.Name) > 100 || strings.IndexFunc(s.Name, unicode.IsControl) >= 0 {
		invalid["name"] = "Enter your name, up to 100 characters."
	}
	a, err := mail.ParseAddress(s.Email)
	if err != nil || a.Address != s.Email || len(s.Email) > 254 || strings.ContainsAny(s.Email, "\r\n") {
		invalid["email"] = "Enter a valid email address."
	}
	if s.Message == "" || utf8.RuneCountInString(s.Message) > 5000 || strings.ContainsRune(s.Message, 0) {
		invalid["message"] = "Enter a message, up to 5,000 characters."
	}
	return s, invalid, values["website"] != ""
}

type Handler struct {
	Notify func(context.Context, Submission) error
}

func reply(w http.ResponseWriter, status int, e appapi.Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(e)
}

// Authentication is supplied by the managed-service wrapper, including the
// private handshake. This handler implements only the app's POST / operation.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.URL.RawPath != "" || r.URL.RawQuery != "" {
		reply(w, 400, appapi.Envelope{Error: "Unsupported contact request."})
		return
	}
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		reply(w, 405, appapi.Envelope{Error: "Use POST to submit a message."})
		return
	}
	fields, err := appapi.ReadFields(w, r)
	if err != nil {
		reply(w, 400, appapi.Envelope{Error: "The contact request is invalid."})
		return
	}
	s, invalid, spam := Validate(fields)
	if len(invalid) != 0 {
		reply(w, 422, appapi.Envelope{Data: Redisplay(fields), Invalid: invalid})
		return
	}
	if !spam {
		if h.Notify == nil || h.Notify(r.Context(), s) != nil {
			reply(w, 503, appapi.Envelope{Data: s.Data(), Error: "We couldn't confirm delivery. Please try again later or contact us directly."})
			return
		}
	}
	reply(w, 200, appapi.Envelope{Redirect: "thanks", Flash: map[string]string{"notice": "Your message was sent."}})
}
