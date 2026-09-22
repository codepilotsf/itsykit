package site

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddAdminPreservesConfigAndAvoidsDuplicates(t *testing.T) {
	for _, admin := range []string{"", `,"admin":null`, `,"admin":{}`, `,"admin":{"emails":null}`, `,"admin":{"emails":["first@example.com"],"custom":true}`} {
		root := fixture(t)
		put(t, root, "_/flea.config", `{"site":{"name":"Keep","custom":9007199254740993},"other":{"enabled":true}`+admin+`}`)
		email, added, err := AddAdmin(root, " New@Example.com ")
		if err != nil || !added || email != "new@example.com" {
			t.Fatal(email, added, err)
		}
		file := filepath.Join(root, "_", "flea.config")
		before, _ := os.ReadFile(file)
		if !strings.Contains(string(before), "9007199254740993") || !strings.Contains(string(before), `"other"`) {
			t.Fatal(string(before))
		}
		var c struct {
			Admin struct {
				Emails []string
				Custom bool
			}
		}
		if err := json.Unmarshal(before, &c); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(admin, "first@example.com") && (len(c.Admin.Emails) != 2 || !c.Admin.Custom) {
			t.Fatal(c)
		}
		if _, added, err := AddAdmin(root, "NEW@example.com"); err != nil || added {
			t.Fatal(added, err)
		}
		after, _ := os.ReadFile(file)
		if string(before) != string(after) {
			t.Fatal("duplicate rewrote configuration")
		}
	}
}

func TestAddAdminRejectsInvalidInputWithoutWrites(t *testing.T) {
	for _, config := range []string{`{`, `null`, `[]`, `{}`, `{"site":{"name":""}}`, `{"site":{"name":"Test"},"admin":[]}`, `{"site":{"name":"Test"},"admin":{"emails":"bad"}}`, `{"site":{"name":"Test"},"admin":{"emails":["bad"]}}`} {
		root := fixture(t)
		put(t, root, "_/flea.config", config)
		if _, _, err := AddAdmin(root, "owner@example.com"); err == nil {
			t.Fatal(config)
		}
		got, _ := os.ReadFile(filepath.Join(root, "_", "flea.config"))
		if string(got) != config {
			t.Fatal("changed invalid config")
		}
	}
	root := fixture(t)
	for _, email := range []string{"", "bad", "Owner <owner@example.com>", "a@example.com\nb@example.com"} {
		if _, _, err := AddAdmin(root, email); err == nil {
			t.Fatal(email)
		}
	}
	for i := 0; i < 20; i++ {
		if _, _, err := AddAdmin(root, fmt.Sprintf("owner%d@example.com", i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := AddAdmin(root, "extra@example.com"); err == nil {
		t.Fatal("allowed over 20 admins")
	}
	if _, added, err := AddAdmin(root, "owner0@example.com"); err != nil || added {
		t.Fatal("duplicate at limit", err)
	}
}

func TestAddAdminRejectsConfigSymlink(t *testing.T) {
	root := fixture(t)
	file := filepath.Join(root, "_", "flea.config")
	outside := filepath.Join(t.TempDir(), "config")
	const body = `{"site":{"name":"Keep"}}`
	if err := os.WriteFile(outside, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, file); err != nil {
		t.Fatal(err)
	}
	if _, _, err := AddAdmin(root, "owner@example.com"); err == nil {
		t.Fatal("followed symlink")
	}
	got, _ := os.ReadFile(outside)
	if string(got) != body {
		t.Fatal("changed external file")
	}
}
