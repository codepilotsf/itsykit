package site

import (
	"fmt"
	"net/mail"
	"strings"
)

// AdminAccess is developer-owned. The browser editor cannot change this list.
type AdminAccess struct {
	Emails          []string `json:"emails"`
	SessionIdleDays int      `json:"session_idle_days,omitempty"`
}

func (a *AdminAccess) validate() error {
	if a.SessionIdleDays == 0 {
		a.SessionIdleDays = 30
	}
	if a.SessionIdleDays < 1 || a.SessionIdleDays > 365 {
		return fmt.Errorf("admin.session_idle_days must be between 1 and 365")
	}
	if len(a.Emails) > 20 {
		return fmt.Errorf("admin.emails supports at most 20 addresses")
	}
	seen := map[string]bool{}
	for i, email := range a.Emails {
		email = strings.ToLower(strings.TrimSpace(email))
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email || len(email) > 254 || seen[email] {
			return fmt.Errorf("admin.emails must contain unique email addresses without display names")
		}
		seen[email] = true
		a.Emails[i] = email
	}
	return nil
}
