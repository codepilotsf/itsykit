package site

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

// validateApps registers the built-in Blog and Contact apps plus any managed
// custom app package under flea/apps/<name>. A custom package is validated at
// load; the running service is still checked by the authenticated handshake at
// startup.
func (p *Project) validateApps(boundary *os.Root) error {
	var mounts []appapi.Mount
	assets := make([]string, 0, len(p.Assets))
	for route := range p.Assets {
		assets = append(assets, route)
	}
	p.appPackages = map[string]appPackage{}
	for name, cfg := range p.Apps {
		if name == "blog" {
			if cfg.Email != (contact.Email{}) {
				return fmt.Errorf("blog does not accept email settings")
			}
			mounts = append(mounts, appapi.Mount{Name: name, Path: cfg.Mount})
			continue
		}
		if name == "contact" {
			if err := cfg.Email.Validate(); err != nil {
				return err
			}
			mounts = append(mounts, appapi.Mount{Name: name, Path: cfg.Mount})
			if p.Pages[cfg.Mount] == nil || p.Pages[cfg.Mount+"/thank-you"] == nil {
				return fmt.Errorf("Contact needs pages at its mount and mount/thank-you")
			}
			continue
		}
		if cfg.Email.To != "" || cfg.Email.From != "" {
			return fmt.Errorf("app %q does not accept email settings", name)
		}
		pkg, err := loadAppPackage(boundary, p.Root, p.RuntimeRoot, name)
		if err != nil {
			return err
		}
		p.appPackages[name] = pkg
		mounts = append(mounts, appapi.Mount{Name: name, Path: cfg.Mount})
		if p.Pages[cfg.Mount] == nil {
			return fmt.Errorf("app %q needs a page at its mount", name)
		}
	}
	return appapi.ValidateMounts(mounts, assets, []string{"/releases"})
}

type runningApp struct {
	mount    appapi.Mount
	config   MiniApp
	manager  *appmanager.Manager
	delivery *contact.Delivery
	pkg      appPackage
}

// samePackage reports whether a running app still matches the loaded package.
// A changed executable or manifest fails closed until the next restart.
func samePackage(running, loaded appPackage) bool {
	return running.Version == loaded.Version && running.Hash == loaded.Hash && running.Size == loaded.Size && running.Executable == loaded.Executable
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
		var spec appmanager.Spec
		var delivery *contact.Delivery
		if name == "contact" {
			delivery = &contact.Delivery{Email: cfg.Email, Sender: sender, Budget: budget, Render: s.RenderEmail}
			spec = appmanager.Spec{Name: name, Bundle: true, Capability: delivery}
		} else {
			pkg, ok := p.appPackages[name]
			if !ok {
				continue
			}
			spec = appmanager.Spec{Name: name, Executable: pkg.Executable, RuntimeRoot: p.RuntimeRoot, Storage: pkg.Storage, Hash: pkg.Hash, Size: pkg.Size, Version: pkg.Version}
		}
		manager, err := appmanager.Start(ctx, spec, s.logger)
		if err != nil && s.logger != nil {
			s.logger.Printf("App %s could not start: %s", name, err)
		}
		s.apps.running[name] = &runningApp{mount: appapi.Mount{Name: name, Path: cfg.Mount}, config: cfg, manager: manager, delivery: delivery, pkg: p.appPackages[name]}
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
			if running != nil && (running.config != cfg || !samePackage(running.pkg, p.appPackages[name])) {
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
