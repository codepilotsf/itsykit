package cli

import (
	"itsy/internal/sitefiles"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"sync"

	"itsy/internal/editor"
	"itsy/internal/editstore"
	"itsy/internal/mailer"
	"itsy/internal/site"
)

// Rebuild authentication state only when local access or mail settings change.
// Sessions persist; pending codes are discarded. Removed owners are revoked.
func devAdmin(root, name string, public *site.Server, logger *log.Logger, preview mailer.Sender) http.Handler {
	var mu sync.Mutex
	var dashboard *editor.Handler
	var emails []string
	var idleDays int
	var settings mailer.Settings
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		adminPath := editor.IsPath(r.URL.Path)
		if !adminPath && r.Header.Get("Cookie") == "" {
			public.ServeHTTP(w, r)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		current, err := public.AdminSettings()
		if err != nil {
			// The site's dev diagnostic remains available even when admin templates
			// or configuration are invalid. This does not bypass authentication.
			public.ServeHTTP(w, r)
			return
		}
		mail, err := mailer.LocalSettings(root)
		if err != nil {
			if !adminPath {
				public.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Check the site's local email settings.", 503)
			return
		}
		if dashboard == nil || !slices.Equal(emails, current.Emails) || idleDays != current.SessionIdleDays || settings != mail {
			var sender mailer.Sender
			if preview != nil {
				sender = preview
			} else if mail.Validate() == nil {
				sender = mailer.NewResend(mail)
			}
			store := &editstore.Store{Root: root, ContentPath: sitefiles.Path(root, filepath.Join("itsy", "snips.json")), Emails: current.Emails, Publish: public.Publish}
			options := editor.Options{Emails: current.Emails, SessionPath: sitefiles.Path(root, filepath.Join("itsy", "runtime", "sessions.json")), SessionIdleDays: current.SessionIdleDays, Name: name, Sender: sender, Backend: store, Local: true, Logger: logger, Render: public.RenderAdmin, Public: public.ServeOwner}
			options.RenderEmail = public.RenderEmail
			options.ReloadScript = public.AdminReload
			if err := editor.PruneSessions(options); err != nil {
				if logger != nil {
					logger.Printf("Admin session storage unavailable: %v", err)
				}
				if !adminPath {
					public.ServeHTTP(w, r)
					return
				}
				http.Error(w, "Check the site's session storage.", 503)
				return
			}
			dashboard = editor.New(options)
			emails, idleDays, settings = current.Emails, current.SessionIdleDays, mail
		}
		if adminPath {
			dashboard.ServeHTTP(w, r)
		} else {
			dashboard.ServePublic(w, r, public)
		}
	})
}
