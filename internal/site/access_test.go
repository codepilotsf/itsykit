package site

import (
	"encoding/json"
	"testing"
)

func TestAdminAllowlistValidation(t *testing.T) {
	for _, emails := range [][]string{{"not an address"}, {"Name <a@example.test>"}, {"A@example.test", "a@example.test"}, make([]string, 21)} {
		a := AdminAccess{Emails: emails}
		if err := a.validate(); err == nil {
			t.Fatal("accepted invalid access list", emails)
		}
	}
	root := fixture(t)
	data, _ := json.Marshal(map[string]any{"site": map[string]string{"name": "Test"}, "admin": AdminAccess{Emails: []string{" Owner@Example.Test "}}})
	put(t, root, "_/flea.config", string(data))
	p, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Admin.Emails[0] != "owner@example.test" {
		t.Fatal(p.Admin)
	}
}
