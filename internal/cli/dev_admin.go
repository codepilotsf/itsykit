package cli

import (
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"sync"

	"flea/internal/editor"
	"flea/internal/editstore"
	"flea/internal/mailer"
	"flea/internal/site"
)

// Rebuild authentication state only when local access or mail settings change.
// Sessions persist; pending codes are discarded. Removed owners are revoked.
func devAdmin(root, name string, public *site.Server, logger *log.Logger) http.Handler {
	var mu sync.Mutex
	var dashboard *editor.Handler
	var emails []string
	var idleDays int
	var settings mailer.Settings
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		current, err := public.AdminSettings()
		if err != nil {
			http.Error(w, "Fix the site configuration before signing in.", 503)
			return
		}
		mail, err := mailer.LocalSettings(root)
		if err != nil {
			http.Error(w, "Check the site's local email settings.", 503)
			return
		}
		if dashboard == nil || !slices.Equal(emails, current.Emails) || idleDays != current.SessionIdleDays || settings != mail {
			var sender mailer.Sender
			if mail.Validate() == nil {
				sender = mailer.NewResend(mail)
			}
			store := &editstore.Store{Root: root, ContentPath: filepath.Join(root, "_", "fleas.json"), Emails: current.Emails, Publish: public.Publish}
			options := editor.Options{Emails: current.Emails, SessionPath: filepath.Join(root, "_", "sessions.json"), SessionIdleDays: current.SessionIdleDays, Name: name, Sender: sender, Backend: store, Local: true, Logger: logger, Render: public.RenderAdmin}
			if err := editor.PruneSessions(options); err != nil {
				if logger != nil {
					logger.Printf("Admin session storage unavailable: %v", err)
				}
				http.Error(w, "Check the site's session storage.", 503)
				return
			}
			dashboard = editor.New(options)
			emails, idleDays, settings = current.Emails, current.SessionIdleDays, mail
		}
		dashboard.ServeHTTP(w, r)
	})
}
