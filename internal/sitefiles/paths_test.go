package sitefiles

import "testing"

func TestProjectBoundaries(t *testing.T) {
	for _, name := range []string{"site", "site/index.html", "site/css/theme.css", "flea", "flea/snips.json", "flea/media/photo.jpg", "flea/media/clip.mp4", "flea/icons/cache.svg", "flea/data/team.json"} {
		if !Included(name) {
			t.Errorf("runtime path excluded: %s", name)
		}
	}
	for _, name := range []string{"index.html", "css/theme.css", "_/snips.json", "flea/runtime", "flea/runtime/sessions.json", "flea/skills/import-brochure/SKILL.md", "site/.secret", "mailers/.env", ".env"} {
		if Included(name) {
			t.Errorf("private or obsolete path admitted: %s", name)
		}
	}
}

func TestConfigurationFilename(t *testing.T) {
	for _, name := range []string{"mailers", "mailers/main.layout.html", "mailers/sign-in.email.html", "mailers/contact.email.txt"} {
		if !Included(name) {
			t.Fatalf("mailer excluded: %s", name)
		}
	}
	if !Included(".flea.config.json") {
		t.Fatal("site configuration excluded")
	}
	if Included("flea/flea.config") {
		t.Fatal("superseded configuration included")
	}
}

func TestSessionsAreNeverRuntimeContent(t *testing.T) {
	for _, name := range []string{"flea/runtime/sessions.json", "flea/runtime/.sessions-lock", "flea/runtime/.sessions-temp"} {
		if Included(name) {
			t.Fatalf("private session state included: %s", name)
		}
	}
}

func TestAgentGuidanceIsNotRuntimeContent(t *testing.T) {
	for _, name := range []string{"AGENTS.md", "checklist.md", "flea/skills", "flea/skills/import-brochure/SKILL.md", "flea/skills/fleacms-site/SKILL.md"} {
		if Included(name) {
			t.Fatalf("agent guidance included in runtime: %s", name)
		}
	}
}
