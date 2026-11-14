package sitefiles

import "testing"

func TestProjectBoundaries(t *testing.T) {
	for _, name := range []string{"site", "site/index.html", "site/css/theme.css", "flea", "itsy/snips.json", "itsy/media/photo.jpg", "itsy/media/clip.mp4", "itsy/icons/cache.svg", "itsy/data/team.json"} {
		if !Included(name) {
			t.Errorf("runtime path excluded: %s", name)
		}
	}
	for _, name := range []string{"index.html", "css/theme.css", "_/snips.json", "itsy/runtime", "itsy/runtime/sessions.json", "itsy/skills/import-brochure/SKILL.md", "site/.secret", "mailers/.env", ".env"} {
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
	if !Included("itsy.config.json") {
		t.Fatal("site configuration excluded")
	}
	if Included("itsy/flea.config") {
		t.Fatal("superseded configuration included")
	}
}

func TestSessionsAreNeverRuntimeContent(t *testing.T) {
	for _, name := range []string{"itsy/runtime/sessions.json", "itsy/runtime/.sessions-lock", "itsy/runtime/.sessions-temp"} {
		if Included(name) {
			t.Fatalf("private session state included: %s", name)
		}
	}
}

func TestAgentGuidanceIsNotRuntimeContent(t *testing.T) {
	for _, name := range []string{"AGENTS.md", "checklist.md", "itsy/skills", "itsy/skills/import-brochure/SKILL.md", ".agents", ".agents/skills/core/SKILL.md", ".agents/site-requirements.md"} {
		if Included(name) {
			t.Fatalf("agent guidance included in runtime: %s", name)
		}
	}
}
