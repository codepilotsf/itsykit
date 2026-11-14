package site

import (
	"context"
	"itsy/internal/snips"
	"net/http"
)

type ownerRenderKey struct{}

// A false result means the transform has already written an error response.
type ownerTransform func([]byte, *snips.Store, string) ([]byte, bool)

// ServeOwner uses the same validated route map and snapshot as public serving.
// Only the authentication layer may call this with an owner-specific transform.
func (s *Server) ServeOwner(w http.ResponseWriter, r *http.Request, transform func([]byte, *snips.Store, string) ([]byte, bool)) {
	s.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ownerRenderKey{}, ownerTransform(transform))))
}
