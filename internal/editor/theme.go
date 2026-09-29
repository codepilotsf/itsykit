package editor

import (
	"encoding/json"
	"flea/internal/theme"
	"io"
	"net/http"
	"strings"
)

func (h *Handler) readTheme(w http.ResponseWriter, r *http.Request) {
	d, e := h.options.Backend.Read()
	if e != nil || d.Theme == nil {
		http.Error(w, "Theme unavailable. Ask your developer to check css/theme.css.", 409)
		return
	}
	fields, e := theme.Fields(*d.Theme)
	if e != nil {
		http.Error(w, "Theme needs one editable :root block. Ask your developer to check css/theme.css.", 409)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Version string                     `json:"version"`
		Fields  []theme.Field              `json:"fields"`
		Scales  map[string]theme.ScaleInfo `json:"scales"`
	}{d.Version, fields, theme.Scales(*d.Theme)})
}
func (h *Handler) saveTheme(w http.ResponseWriter, r *http.Request) {
	var edit struct {
		Version string                     `json:"version"`
		Changes map[string]string          `json:"changes"`
		Scales  map[string]theme.ScaleEdit `json:"scales"`
	}
	d := json.NewDecoder(strings.NewReader(r.PostForm.Get("edit")))
	d.DisallowUnknownFields()
	if e := d.Decode(&edit); e != nil || d.Decode(new(any)) != io.EOF || edit.Version == "" || len(edit.Changes)+len(edit.Scales) == 0 || len(edit.Changes) > 200 || len(edit.Scales) > 2 {
		http.Error(w, "Invalid theme changes.", 400)
		return
	}
	if e := h.options.Backend.Save(Update{Version: edit.Version, Theme: edit.Changes, Scales: edit.Scales}); e != nil {
		if h.options.Logger != nil {
			h.options.Logger.Printf("Theme save failed: %q", e.Error())
		}
		http.Error(w, "Could not save. The theme may have changed. Your preview is still here; copy your changes before reloading.", 409)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
