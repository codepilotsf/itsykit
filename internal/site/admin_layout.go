package site

import (
	"fmt"
	"html/template"
)

// AdminSettings returns a copy of the current validated access list.
func (s *Server) AdminSettings() (AdminAccess, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.err != nil {
		return AdminAccess{}, s.err
	}
	a := s.project.Admin
	a.Emails = append([]string(nil), a.Emails...)
	return a, nil
}

// Admin templates are site-owned Flea HTML, compiled with the normal layout.
// They are excluded from public routes and can only be selected by editor state.
func (p *Project) buildAdminLayout() error {
	p.adminTemplates = map[string]*template.Template{}
	for _, name := range []string{"index", "code", "setup", "error"} {
		file := "admin/" + name + ".html"
		if _, ok := p.files[file]; !ok {
			continue
		}
		t, err := p.buildTemplate(file, true)
		if err != nil {
			return err
		}
		if _, err := p.renderWithBody(t, "/admin", map[string]any{}); err != nil {
			return err
		}
		p.adminTemplates[name] = t
	}
	return nil
}

// RenderAdmin uses the current immutable snapshot, including live template edits.
func (s *Server) RenderAdmin(mode string, data map[string]any) ([]byte, error) {
	s.mu.RLock()
	p, err := s.project, s.err
	s.mu.RUnlock()
	if err != nil {
		return nil, err
	}
	names := map[string]string{"login": "index", "verify": "code", "setup": "setup", "error": "error"}
	name, ok := names[mode]
	if !ok {
		return nil, fmt.Errorf("unknown admin page")
	}
	t := p.adminTemplates[name]
	if t == nil {
		return nil, fmt.Errorf("missing admin/%s.html; copy the admin directory from a fresh Flea site", name)
	}
	return p.renderWithBody(t, "/admin", data)
}

// AdminReload supplies trusted development JavaScript to the editor's CSP-aware renderer.
func (s *Server) AdminReload() string {
	if s.production {
		return ""
	}
	s.mu.RLock()
	revision := s.revision
	s.mu.RUnlock()
	script := reloadScript(revision)
	return string(script[len("<script>") : len(script)-len("</script>")])
}
