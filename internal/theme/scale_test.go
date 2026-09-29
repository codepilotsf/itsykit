package theme

import (
	"math"
	"os"
	"strings"
	"testing"
)

func TestScaleMath(t *testing.T) {
	for _, test := range []struct {
		s    Scale
		step int
		want string
	}{
		{defaultScale("text"), 0, "clamp(0.875rem, 0.8125rem + 0.3125vw, 1.125rem)"},
		{Scale{320, 1600, 10, 100, 1.25, 1.414}, 0, "clamp(0.625rem, -0.7813rem + 7.031vw, 6.25rem)"},
		{Scale{320, 1600, 100, 10, 1.25, 1.414}, 0, "clamp(0.625rem, 7.656rem - 7.031vw, 6.25rem)"},
	} {
		if got := Clamp(test.s, test.step); got != test.want {
			t.Fatalf("%s != %s", got, test.want)
		}
	}
	for _, s := range []Scale{{320, 320, 14, 18, 1.25, 1.414}, {320, 1600, 0, 18, 1.25, 1.414}, {320, 1600, 14, 18, math.NaN(), 1.414}} {
		if _, err := Generate("text", s); err == nil {
			t.Fatal("invalid accepted")
		}
	}
}
func TestScaleRoundTripAndCustom(t *testing.T) {
	b, err := os.ReadFile("../../templates/site/css/theme.css")
	if err != nil {
		t.Fatal(err)
	}
	css := string(b)
	info := Scales(css)
	if len(info) != 2 || info["text"].Custom || info["space"].Custom {
		t.Fatalf("starter: %#v", info)
	}
	s := info["text"].Settings
	s.MaxBase = 22
	next, err := ApplyScales(css, map[string]string{"--brand-1": "#123456"}, map[string]ScaleEdit{"text": {Settings: s}})
	if err != nil {
		t.Fatal(err)
	}
	if Scales(next)["text"].Custom || Scales(next)["text"].Settings != s || !strings.Contains(next, "Max base size: 22px") {
		t.Fatal("settings not retained")
	}
	if !strings.Contains(next, "--space-sm: clamp(0.75rem, 0.6563rem + 0.4688vw, 1.125rem)") {
		t.Fatal("unrelated spacing changed")
	}
	custom := strings.Replace(css, "--text-base: clamp(0.875rem, 0.8125rem + 0.3125vw, 1.125rem)", "--text-base: 2rem", 1)
	if !Scales(custom)["text"].Custom {
		t.Fatal("custom not detected")
	}
	edits := map[string]ScaleEdit{"text": {Settings: s}}
	if _, err := ApplyScales(custom, nil, edits); err == nil {
		t.Fatal("custom overwritten")
	}
	edits["text"] = ScaleEdit{Settings: s, Replace: true}
	next, err = ApplyScales(custom, nil, edits)
	if err != nil || Scales(next)["text"].Custom {
		t.Fatal(err)
	}
	if _, err := ApplyScales(css, nil, map[string]ScaleEdit{"evil": {Settings: s}}); err == nil {
		t.Fatal("unknown scale")
	}
	if _, err := ApplyScales(css, map[string]string{"--text-base": "2rem"}, edits); err == nil {
		t.Fatal("conflicting inputs")
	}
}
