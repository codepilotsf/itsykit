package editor

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"itsy/internal/media"
	"itsy/internal/snips"
	"itsy/internal/video"
)

// MediaValidationError contains only a safe validation explanation, never
// filesystem paths, provider errors or other operational details.
type MediaValidationError struct{ Message string }

var mediaSlot = make(chan struct{}, 1)

func (e *MediaValidationError) Error() string { return e.Message }

// Called only after the ordinary Origin, session, CSRF and body-size checks.
func (h *Handler) savePage(w http.ResponseWriter, r *http.Request) {
	var edit PageEdit
	d := json.NewDecoder(strings.NewReader(r.PostForm.Get("edit")))
	d.DisallowUnknownFields()
	if err := d.Decode(&edit); err != nil {
		http.Error(w, "Invalid page changes. Reload and try again.", 400)
		return
	}
	if d.Decode(new(any)) != io.EOF || len(edit.Changes) > 200 || len(edit.Media) > 100 || len(r.PostForm.Get("edit")) > 1<<20 || edit.Version == "" || edit.Route == "" {
		http.Error(w, "Invalid page changes. Reload and try again.", 400)
		return
	}
	allowedFiles := map[string]bool{}
	seen := map[string]bool{}
	for i := range edit.Media {
		change := &edit.Media[i]
		if !snips.ValidID(change.ID) || seen[change.ID] {
			http.Error(w, "Invalid media changes.", 400)
			return
		}
		seen[change.ID] = true
		fileLimit := int64(video.MaxFile)
		if change.Library {
			fileLimit = media.MaxFile
		}
		for _, part := range []struct {
			name   string
			target *[]byte
			limit  int64
		}{
			{"media-" + change.ID, &change.Data, fileLimit},
			{"poster-" + change.ID, &change.PosterData, media.MaxFile},
		} {
			allowedFiles[part.name] = true
			if r.MultipartForm == nil || len(r.MultipartForm.File[part.name]) == 0 {
				continue
			}
			files := r.MultipartForm.File[part.name]
			if len(files) != 1 || files[0].Size <= 0 || files[0].Size > part.limit {
				http.Error(w, "Invalid or oversized media upload.", 400)
				return
			}
			f, err := files[0].Open()
			if err != nil {
				http.Error(w, "Could not read upload.", 400)
				return
			}
			*part.target, err = io.ReadAll(io.LimitReader(f, part.limit+1))
			f.Close()
			if err != nil || int64(len(*part.target)) > part.limit {
				http.Error(w, "Could not read upload.", 400)
				return
			}
		}
	}
	if r.MultipartForm != nil {
		for name := range r.MultipartForm.File {
			if !allowedFiles[name] {
				http.Error(w, "Unexpected media upload.", 400)
				return
			}
		}
	}
	var result PageResult
	if err := h.options.Backend.Save(Update{Page: &edit, Result: &result}); err != nil {
		if h.options.Logger != nil {
			h.options.Logger.Printf("Page save failed: %q", err.Error())
		}
		var validation *MediaValidationError
		if errors.As(err, &validation) {
			http.Error(w, validation.Message, 400)
			return
		}
		http.Error(w, "Could not save. The page may have changed. Your draft is still here; copy it before reloading.", 409)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
