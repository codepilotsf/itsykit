package site

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"flea/internal/appapi"
	"flea/internal/appmanager"
	"flea/internal/contact"
	"flea/internal/mailer"
)

type MiniApp struct {
	Mount string        `json:"mount"`
	Email contact.Email `json:"email"`
}

func (p *Project) validateApps() error {
	var mounts []appapi.Mount
	assets := make([]string, 0, len(p.Assets))
	for route := range p.Assets {
		assets = append(assets, route)
	}
	for name, cfg := range p.Apps {
		if name != "contact" && name != "blog" {
			return fmt.Errorf("app %q is not included in this build", name)
		}
		if name == "blog" {
			if cfg.Email != (contact.Email{}) {
				return fmt.Errorf("blog does not accept email settings")
			}
			mounts = append(mounts, appapi.Mount{Name: name, Path: cfg.Mount})
			continue
		}
		if err := cfg.Email.Validate(); err != nil {
			return err
		}
		mounts = append(mounts, appapi.Mount{Name: name, Path: cfg.Mount})
		if p.Pages[cfg.Mount] == nil || p.Pages[cfg.Mount+"/thank-you"] == nil {
			return fmt.Errorf("Contact needs pages at its mount and mount/thank-you")
		}
	}
	return appapi.ValidateMounts(mounts, assets, []string{"/releases"})
}

type runningApp struct {
	mount    appapi.Mount
	config   MiniApp
	manager  *appmanager.Manager
	delivery *contact.Delivery
}
type appState struct {
	sync.RWMutex
	running map[string]*runningApp
	flashes flashStore
}

// StartApps is called only by an actual dev/host run, never read-only loading.
// An unavailable app is isolated to its routes; no legacy-handler fallback.
func (s *Server) StartApps(ctx context.Context, sender mailer.Sender, budget *contact.Budget) func() {
	s.mu.RLock()
	p := s.project
	s.mu.RUnlock()
	s.apps.Lock()
	defer s.apps.Unlock()
	if s.apps.running != nil {
		return func() {}
	}
	s.apps.running = map[string]*runningApp{}
	for name, cfg := range p.Apps {
		if name == "blog" {
			continue
		}
		delivery := &contact.Delivery{Email: cfg.Email, Sender: sender, Budget: budget, Render: s.RenderEmail}
		manager, err := appmanager.Start(ctx, delivery, s.logger)
		if err != nil && s.logger != nil {
			s.logger.Printf("Contact could not start: %s", err)
		}
		s.apps.running[name] = &runningApp{mount: appapi.Mount{Name: name, Path: cfg.Mount}, config: cfg, manager: manager, delivery: delivery}
	}
	return func() {
		s.apps.Lock()
		running := s.apps.running
		s.apps.running = nil
		s.apps.Unlock()
		for _, app := range running {
			if app.manager != nil {
				app.manager.Close()
			}
		}
	}
}

func (s *Server) appFor(p *Project, route string) (appapi.Mount, *runningApp, bool) {
	for name, cfg := range p.Apps {
		if name == "blog" {
			continue
		}
		if route == cfg.Mount || strings.HasPrefix(route, cfg.Mount+"/") {
			s.apps.RLock()
			running := s.apps.running[name]
			s.apps.RUnlock()
			if running != nil && running.config != cfg {
				running = nil
			}
			return appapi.Mount{Name: name, Path: cfg.Mount}, running, true
		}
	}
	return appapi.Mount{}, nil, false
}

func decodeApps(raw json.RawMessage) (map[string]MiniApp, error) {
	if raw == nil {
		return nil, nil
	}
	if string(raw) == "null" {
		return nil, fmt.Errorf("apps must be an object")
	}
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	var configs map[string]MiniApp
	if err := d.Decode(&configs); err != nil {
		return nil, fmt.Errorf("apps must contain supported app settings")
	}
	return configs, nil
}
