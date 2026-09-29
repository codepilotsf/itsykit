package theme

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Scale settings stay in theme.css with their generated declarations.
type Scale struct {
	MinViewport float64 `json:"minViewport"`
	MaxViewport float64 `json:"maxViewport"`
	MinBase     float64 `json:"minBase"`
	MaxBase     float64 `json:"maxBase"`
	MinRatio    float64 `json:"minRatio"`
	MaxRatio    float64 `json:"maxRatio"`
}
type ScaleInfo struct {
	Settings Scale `json:"settings"`
	Custom   bool  `json:"custom"`
}
type ScaleEdit struct {
	Settings Scale `json:"settings"`
	Replace  bool  `json:"replace"`
}

func scaleNames(kind string) []string {
	if kind == "text" {
		return []string{"sm", "base", "lg", "xl", "2xl", "3xl", "4xl"}
	}
	if kind == "space" {
		return []string{"3xs", "2xs", "xs", "sm", "md", "lg", "xl", "2xl", "3xl", "4xl"}
	}
	return nil
}
func defaultScale(kind string) Scale {
	if kind == "text" {
		return Scale{320, 1600, 14, 18, 1.25, 1.414}
	}
	return Scale{320, 1600, 12, 18, 1.5, 1.667}
}
func (s Scale) Validate() error {
	for _, v := range []float64{s.MinViewport, s.MaxViewport, s.MinBase, s.MaxBase, s.MinRatio, s.MaxRatio} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("invalid scale number")
		}
	}
	if s.MinViewport < 100 || s.MaxViewport > 10000 || s.MaxViewport-s.MinViewport < 1 || s.MinBase < 1 || s.MaxBase < 1 || s.MinBase > 200 || s.MaxBase > 200 || s.MinRatio < 1 || s.MaxRatio < 1 || s.MinRatio > 2 || s.MaxRatio > 2 {
		return fmt.Errorf("invalid scale settings")
	}
	return nil
}
func number(v float64) string {
	if v == 0 {
		return "0"
	}
	factor := math.Pow(10, 3-math.Floor(math.Log10(math.Abs(v))))
	rounded := math.Round(v*factor) / factor
	if rounded == 0 {
		return "0"
	}
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}
func Clamp(s Scale, step int) string {
	a, b := s.MinBase*math.Pow(s.MinRatio, float64(step)), s.MaxBase*math.Pow(s.MaxRatio, float64(step))
	slope := (b - a) / (s.MaxViewport - s.MinViewport)
	intercept := (a - slope*s.MinViewport) / 16
	sign := " + "
	if slope < 0 {
		sign = " - "
	}
	return "clamp(" + number(math.Min(a, b)/16) + "rem, " + number(intercept) + "rem" + sign + number(math.Abs(slope)*100) + "vw, " + number(math.Max(a, b)/16) + "rem)"
}
func Generate(kind string, s Scale) (map[string]string, error) {
	names := scaleNames(kind)
	if names == nil {
		return nil, fmt.Errorf("unknown scale")
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	shift := 1
	if kind == "space" {
		shift = 3
	}
	values := map[string]string{}
	for i, name := range names {
		values["--"+kind+"-"+name] = Clamp(s, i-shift)
	}
	return values, nil
}
func scaleComment(kind string) *regexp.Regexp {
	title := "Text"
	if kind == "space" {
		title = "Space"
	}
	return regexp.MustCompile(`/\* Fluid ` + title + ` Scale[^*]*\*/`)
}

var numeric = regexp.MustCompile(`[-+]?(?:\d*\.)?\d+(?:e[-+]?\d+)?`)

func matchingValue(a, b string) bool {
	x, y := numeric.FindAllString(a, -1), numeric.FindAllString(b, -1)
	if len(x) != 4 || len(y) != 4 || numeric.ReplaceAllString(a, "#") != numeric.ReplaceAllString(b, "#") {
		return false
	}
	for i := range x {
		v, _ := strconv.ParseFloat(x[i], 64)
		w, _ := strconv.ParseFloat(y[i], 64)
		if math.Abs(v-w) > math.Max(0.00001, math.Abs(w)*0.003) {
			return false
		}
	}
	return true
}
func Scales(css string) map[string]ScaleInfo {
	fields, err := Fields(css)
	if err != nil {
		return nil
	}
	known := map[string]string{}
	for _, f := range fields {
		known[f.Name] = f.Value
	}
	result := map[string]ScaleInfo{}
	for _, kind := range []string{"text", "space"} {
		present := true
		for _, name := range scaleNames(kind) {
			if _, ok := known["--"+kind+"-"+name]; !ok {
				present = false
			}
		}
		if !present {
			continue
		}
		s := defaultScale(kind)
		comments := scaleComment(kind).FindAllString(css, -1)
		custom := true
		if len(comments) == 1 {
			base := "base"
			if kind == "space" {
				base = "sm"
			}
			keys := []string{"Min viewport", "Max viewport", "Min " + base + " size", "Max " + base + " size", "Min ratio", "Max ratio"}
			values := []*float64{&s.MinViewport, &s.MaxViewport, &s.MinBase, &s.MaxBase, &s.MinRatio, &s.MaxRatio}
			custom = false
			for i, key := range keys {
				m := regexp.MustCompile(regexp.QuoteMeta(key) + `: ([0-9.]+)`).FindStringSubmatch(comments[0])
				if len(m) != 2 {
					custom = true
					break
				}
				*values[i], err = strconv.ParseFloat(m[1], 64)
				if err != nil {
					custom = true
					break
				}
			}
			if s.Validate() != nil {
				custom = true
				s = defaultScale(kind)
			}
			generated, _ := Generate(kind, s)
			for name, value := range generated {
				if !matchingValue(known[name], value) {
					custom = true
				}
			}
		}
		result[kind] = ScaleInfo{s, custom}
	}
	return result
}
func ApplyScales(css string, changes map[string]string, edits map[string]ScaleEdit) (string, error) {
	all := map[string]string{}
	for k, v := range changes {
		all[k] = v
	}
	info := Scales(css)
	for kind, edit := range edits {
		if len(scaleComment(kind).FindAllStringIndex(css, -1)) > 1 {
			return "", fmt.Errorf("duplicate scale settings comments")
		}
		current, ok := info[kind]
		if !ok || (current.Custom && !edit.Replace) {
			return "", fmt.Errorf("custom scale requires explicit replacement")
		}
		values, err := Generate(kind, edit.Settings)
		if err != nil {
			return "", err
		}
		for name, value := range values {
			if _, ok := changes[name]; ok {
				return "", fmt.Errorf("conflicting scale changes")
			}
			all[name] = value
		}
	}
	next, err := Apply(css, all)
	if err != nil {
		return "", err
	}
	for _, kind := range []string{"text", "space"} {
		edit, ok := edits[kind]
		if !ok {
			continue
		}
		s := edit.Settings
		title, base := "Text", "base"
		if kind == "space" {
			title, base = "Space", "sm"
		}
		comment := fmt.Sprintf("/* Fluid %s Scale\n     Min viewport: %gpx\n     Max viewport: %gpx\n     Min %s size: %gpx\n     Max %s size: %gpx\n     Min ratio: %g\n     Max ratio: %g\n  */", title, s.MinViewport, s.MaxViewport, base, s.MinBase, base, s.MaxBase, s.MinRatio, s.MaxRatio)
		matches := scaleComment(kind).FindAllStringIndex(next, -1)
		if len(matches) == 1 {
			m := matches[0]
			next = next[:m[0]] + comment + next[m[1]:]
		} else {
			next = comment + "\n" + next
		}
	}
	if len(next) > MaxBytes || strings.Contains(next, "\x00") {
		return "", fmt.Errorf("invalid theme")
	}
	return next, nil
}
