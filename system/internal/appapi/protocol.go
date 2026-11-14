// Package appapi implements the internal ItsyWeb App Protocol foundation.
// It does not discover, launch, or publicly mount services.
package appapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"itsy/internal/formguard"
)

// Version is the current ItsyWeb App Protocol. It was 1 before the app-owned
// session envelope and the managed launch contract were added; v1 services are
// rejected explicitly rather than adapted.
const Version = 2
const FormLimit = 32 << 10
const ResponseLimit = 1 << 20
const HandshakeLimit = 4 << 10

// App-owned session bridge. ItsyWeb reads its own per-site/per-app cookie and
// forwards the opaque token as trusted transport metadata; it never forwards a
// browser Cookie or Authorization header. Token length and lifetime are bounded.
const (
	SessionHeader   = "X-Flea-Session"
	SessionTokenMin = 16
	SessionTokenMax = 512
	SessionMaxAge   = 30 * 24 * 60 * 60 // seconds
)

// SessionDirective is the validated instruction an app gives ItsyWeb for its
// browser cookie. Present is false when the app omitted the key entirely.
type SessionDirective struct {
	Present bool
	Clear   bool
	Token   string
	MaxAge  int
}

// ValidSessionToken bounds token length and restricts it to cookie-safe
// printable ASCII so it can travel in a header and a Set-Cookie value unchanged.
func ValidSessionToken(token string) bool {
	if len(token) < SessionTokenMin || len(token) > SessionTokenMax {
		return false
	}
	for i := 0; i < len(token); i++ {
		c := token[i]
		if c < 0x21 || c > 0x7e {
			return false
		}
		switch c {
		case ';', ',', '"', '\\', ' ':
			return false
		}
	}
	return true
}

type Identity struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Protocol int    `json:"protocol"`
}

type Envelope struct {
	Data     map[string]any    `json:"data,omitempty"`
	Invalid  map[string]string `json:"invalid,omitempty"`
	Error    string            `json:"error,omitempty"`
	Flash    map[string]string `json:"flash,omitempty"`
	Redirect string            `json:"redirect,omitempty"`
	Session  json.RawMessage   `json:"session,omitempty"`
}

// SessionDirective validates the app's session instruction. Omission leaves the
// cookie unchanged; an explicit null clears it; an object must contain exactly a
// valid opaque token and a bounded integer max_age in seconds.
func (e Envelope) SessionDirective() (SessionDirective, error) {
	raw := bytes.TrimSpace(e.Session)
	if len(raw) == 0 {
		return SessionDirective{}, nil
	}
	if string(raw) == "null" {
		return SessionDirective{Present: true, Clear: true}, nil
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := checkJSON(d, 0); err != nil {
		return SessionDirective{}, fmt.Errorf("invalid session object")
	}
	if _, err := d.Token(); err != io.EOF {
		return SessionDirective{}, fmt.Errorf("invalid session object")
	}
	// Field names are exact and case-sensitive: the documented contract is
	// `token` and `max_age`, and no aliases are accepted.
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil || len(body) != 2 {
		return SessionDirective{}, fmt.Errorf("invalid session object")
	}
	tokenRaw, ok := body["token"]
	if !ok {
		return SessionDirective{}, fmt.Errorf("invalid session object")
	}
	ageRaw, ok := body["max_age"]
	if !ok {
		return SessionDirective{}, fmt.Errorf("invalid session object")
	}
	var token string
	if len(tokenRaw) == 0 || tokenRaw[0] != '"' || json.Unmarshal(tokenRaw, &token) != nil || !ValidSessionToken(token) {
		return SessionDirective{}, fmt.Errorf("invalid session token")
	}
	// max_age must be a bare JSON integer, never a quoted string.
	if len(ageRaw) == 0 || !bareInteger.Match(ageRaw) {
		return SessionDirective{}, fmt.Errorf("invalid session lifetime")
	}
	age, err := strconv.ParseInt(string(ageRaw), 10, 32)
	if err != nil || age < 1 || age > SessionMaxAge {
		return SessionDirective{}, fmt.Errorf("invalid session lifetime")
	}
	return SessionDirective{Present: true, Token: token, MaxAge: int(age)}, nil
}

var bareInteger = regexp.MustCompile(`^-?[0-9]+$`)

// Values is a fresh top-level map. Nothing is merged into the shared Project.
func (e Envelope) Values() map[string]any {
	invalid := make(map[string]any, len(e.Invalid))
	for k, v := range e.Invalid {
		invalid[k] = v
	}
	flash := make(map[string]any, len(e.Flash))
	for k, v := range e.Flash {
		flash[k] = v
	}
	return map[string]any{"data": e.Data, "invalid": invalid, "error": e.Error, "flash": flash}
}

type Mount struct{ Name, Path string }

var appName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ValidName reserves every built-in template namespace so an app name can never
// shadow a core template value. The built-in Blog global is guarded separately:
// apps.blog selects the bundled Blog app, so no custom package can take that
// name, and blog is a valid app name for mount validation.
func ValidName(name string) bool {
	return len(name) <= 64 && appName.MatchString(name) && !strings.Contains(" app page snip data media videos admin csrf ", " "+name+" ")
}

func publicPath(value string) bool {
	if value == "" || !strings.HasPrefix(value, "/") || path.Clean(value) != value || strings.ContainsAny(value, "\\%?#") || strings.IndexFunc(value, unicode.IsSpace) >= 0 || strings.IndexFunc(value, unicode.IsControl) >= 0 || !utf8.ValidString(value) {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if strings.HasPrefix(part, ".") || strings.HasPrefix(part, "_") || strings.Contains(strings.ToLower(part), ".html") {
			return false
		}
	}
	return true
}

func under(value, mount string) bool { return value == mount || strings.HasPrefix(value, mount+"/") }

// ValidateMounts permits site-owned pages beneath a mount, but never assets or
// core routes. core contains any additional host-owned routes (e.g. downloads).
func ValidateMounts(mounts []Mount, assets, core []string) error {
	names := map[string]bool{}
	for i, m := range mounts {
		if !ValidName(m.Name) || names[m.Name] {
			return fmt.Errorf("app names must be unique and must not use a reserved template namespace")
		}
		names[m.Name] = true
		if m.Path == "/" || !publicPath(m.Path) {
			return fmt.Errorf("app mount must be a canonical public path below the site root")
		}
		for _, reserved := range append([]string{"/admin", "/_flea"}, core...) {
			if under(m.Path, reserved) || under(reserved, m.Path) {
				return fmt.Errorf("app mount conflicts with a core route")
			}
		}
		for _, asset := range assets {
			if under(asset, m.Path) || under(m.Path, asset) {
				return fmt.Errorf("app mount conflicts with an asset")
			}
		}
		for _, previous := range mounts[:i] {
			if under(m.Path, previous.Path) || under(previous.Path, m.Path) {
				return fmt.Errorf("app mounts overlap")
			}
		}
	}
	return nil
}

func (m Mount) ServicePath(public string) (string, error) {
	if err := ValidateMounts([]Mount{m}, nil, nil); err != nil {
		return "", err
	}
	if !publicPath(public) || !under(public, m.Path) {
		return "", fmt.Errorf("route is outside the app mount")
	}
	if public == m.Path {
		return "/", nil
	}
	return strings.TrimPrefix(public, m.Path), nil
}

// Redirects are descendants of the configured mount, never origin-relative or
// external URLs. Encoded traversal, query strings and fragments are rejected.
func (m Mount) Redirect(destination string) (string, error) {
	if destination == "" || strings.HasPrefix(destination, "/") {
		return "", fmt.Errorf("redirect must be mount-relative")
	}
	if !publicPath("/" + destination) {
		return "", fmt.Errorf("invalid redirect")
	}
	if err := ValidateMounts([]Mount{m}, nil, nil); err != nil {
		return "", err
	}
	return m.Path + "/" + destination, nil
}

type RequestError struct {
	Status  int
	Message string
}

func (e *RequestError) Error() string { return e.Message }

// DecodeForm keeps query values separate, removes only ItsyWeb's verified field,
// and otherwise preserves literal names, empty strings, and repeated values.
func DecodeForm(w http.ResponseWriter, r *http.Request, guard *formguard.Guard) (map[string]any, error) {
	if r.Method != http.MethodPost {
		return nil, &RequestError{405, "Use POST to submit this form."}
	}
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || kind != "application/x-www-form-urlencoded" {
		return nil, &RequestError{415, "Use a standard form without file uploads."}
	}
	r.Body = http.MaxBytesReader(w, r.Body, FormLimit)
	// Never trust a previously merged Form supplied by another layer.
	r.Form, r.PostForm = nil, nil
	if err := r.ParseForm(); err != nil {
		return nil, &RequestError{400, "The form is too large or invalid."}
	}
	if guard == nil || guard.Verify(r, r.PostForm) != nil {
		return nil, &RequestError{403, formguard.ErrInvalid.Error()}
	}
	fields := make(map[string]any, len(r.PostForm))
	for name, values := range r.PostForm {
		if name == formguard.Field {
			continue
		}
		if !utf8.ValidString(name) {
			return nil, &RequestError{400, "Form fields must use valid UTF-8."}
		}
		for _, value := range values {
			if !utf8.ValidString(value) {
				return nil, &RequestError{400, "Form fields must use valid UTF-8."}
			}
		}
		if len(values) == 1 {
			fields[name] = values[0]
		} else {
			fields[name] = append([]string(nil), values...)
		}
	}
	return fields, nil
}

// Walk once to reject duplicate keys and excessive nesting before decoding
// typed protocol values. encoding/json alone silently accepts duplicate keys.
func checkJSON(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("JSON nesting exceeds the limit")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return fmt.Errorf("invalid JSON")
	}
	seen := map[string]bool{}
	for d.More() {
		if delim == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return fmt.Errorf("duplicate JSON field")
			}
			seen[name] = true
		}
		if err := checkJSON(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}

func decode(body []byte, target any) error {
	trimmed := strings.TrimSpace(string(body))
	if !utf8.Valid(body) || len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("expected a JSON object")
	}
	d := json.NewDecoder(strings.NewReader(string(body)))
	d.UseNumber()
	if err := checkJSON(d, 0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected one JSON object")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}
	allowed := map[string]bool{}
	switch target.(type) {
	case *Identity:
		for _, key := range []string{"name", "version", "protocol"} {
			allowed[key] = true
		}
		if len(raw) != 3 {
			return fmt.Errorf("missing handshake fields")
		}
	case *Envelope:
		for _, key := range []string{"data", "invalid", "error", "flash", "redirect", "session"} {
			allowed[key] = true
		}
	default:
		return fmt.Errorf("unsupported protocol object")
	}
	for key, value := range raw {
		if !allowed[key] {
			return fmt.Errorf("invalid protocol field")
		}
		// An explicit null is the documented way to clear an app session, so it
		// is validated by SessionDirective rather than rejected here.
		if key == "session" {
			continue
		}
		if string(value) == "null" {
			return fmt.Errorf("invalid protocol field")
		}
		if key == "invalid" || key == "flash" {
			var messages map[string]json.RawMessage
			if err := json.Unmarshal(value, &messages); err != nil {
				return err
			}
			for _, message := range messages {
				if len(message) == 0 || message[0] != '"' {
					return fmt.Errorf("messages must be strings")
				}
			}
		}
	}
	d = json.NewDecoder(strings.NewReader(string(body)))
	d.UseNumber()
	d.DisallowUnknownFields()
	return d.Decode(target)
}
