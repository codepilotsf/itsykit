package site

import (
	"fmt"
	"net/mail"
	"strings"
)

// AdminAccess is developer-owned. The browser editor cannot change this list.
type AdminAccess struct {
	Emails []string `json:"emails"`
}

func (a *AdminAccess) validate() error {
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
