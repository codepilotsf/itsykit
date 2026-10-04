// Package appapi implements the internal Flea App Protocol foundation.
// It does not discover, launch, or publicly mount services.
package appapi

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"flea/internal/formguard"
)

const Version = 1
const FormLimit = 32 << 10
const ResponseLimit = 1 << 20
const HandshakeLimit = 4 << 10

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
}

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

// DecodeForm keeps query values separate, removes only Flea's verified field,
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
		for _, key := range []string{"data", "invalid", "error", "flash", "redirect"} {
			allowed[key] = true
		}
	default:
		return fmt.Errorf("unsupported protocol object")
	}
	for key, value := range raw {
		if !allowed[key] || string(value) == "null" {
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
