package site

import (
	"encoding/json"
	"testing"
)

func TestAdminAllowlistValidation(t *testing.T) {
	for _, days := range []int{-1, 366} {
		a := AdminAccess{SessionIdleDays: days}
		if a.validate() == nil {
			t.Fatal("accepted invalid idle window", days)
		}
	}
	for _, days := range []int{0, 1, 30, 365} {
		a := AdminAccess{SessionIdleDays: days}
		if err := a.validate(); err != nil {
			t.Fatal(err)
		}
		if days == 0 && a.SessionIdleDays != 30 {
			t.Fatal("missing default")
		}
	}
	for _, emails := range [][]string{{"not an address"}, {"Name <a@example.test>"}, {"A@example.test", "a@example.test"}, make([]string, 21)} {
		a := AdminAccess{Emails: emails}
		if err := a.validate(); err == nil {
			t.Fatal("accepted invalid access list", emails)
		}
	}
	root := fixture(t)
	data, _ := json.Marshal(map[string]any{"app": map[string]string{"site_name": "Test"}, "admin": AdminAccess{Emails: []string{" Owner@Example.Test "}}})
	put(t, root, "itsy.config.json", string(data))
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Admin.Emails[0] != "owner@example.test" {
		t.Fatal(p.Admin)
	}
}
