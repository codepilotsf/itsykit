package appapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"unicode/utf8"
)

// ReadFields decodes the JSON service side of the flat form contract, rejecting
// duplicate keys rather than silently choosing a value. No request envelope.
func ReadFields(w http.ResponseWriter, r *http.Request) (map[string]any, error) {
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || kind != "application/json" {
		return nil, fmt.Errorf("expected JSON")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 6*FormLimit)
	body, err := io.ReadAll(r.Body)
	if err != nil || !utf8.Valid(body) {
		return nil, fmt.Errorf("invalid request")
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 || body[0] != '{' {
		return nil, fmt.Errorf("expected an object")
	}
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	if err := checkJSON(d, 0); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("expected one object")
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, err
	}
	for name, value := range fields {
		switch value := value.(type) {
		case string:
		case []any:
			if len(value) < 2 {
				return nil, fmt.Errorf("repeated fields need at least two values")
			}
			for _, item := range value {
				if _, ok := item.(string); !ok {
					return nil, fmt.Errorf("field values must be strings")
				}
			}
		default:
			return nil, fmt.Errorf("field values must be strings")
		}
		if name == "_csrf" {
			return nil, fmt.Errorf("CSRF belongs to ItsyKit")
		}
	}
	return fields, nil
}
