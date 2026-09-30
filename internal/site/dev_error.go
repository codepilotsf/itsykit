package site

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"errors"
	"html/template"
	"net"
	"net/http"
	"strings"

	"flea/internal/devaccess"
)

//go:embed dev_error.html
var devErrorHTML string
var devErrorTemplate = template.Must(template.New("dev-error").Parse(devErrorHTML))

func localDiagnosticRequest(r *http.Request) bool {
	if devaccess.LAN(r) {
		return true
	}
	host := r.Host
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	if !strings.EqualFold(host, "localhost") && (ip == nil || !ip.IsLoopback()) {
		return false
	}
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	return err == nil && net.ParseIP(peer) != nil && net.ParseIP(peer).IsLoopback()
}

type devErrorView struct {
	Code                                             int
	Title, Summary, Detail, Hint, Path, Method, File string
	Line                                             int
	Lines                                            []sourceLine
}

func serveDevError(w http.ResponseWriter, r *http.Request, p *Project, code int, revision uint64, cause error) {
	v := devErrorView{Code: code, Title: "Application error", Summary: "Flea found a problem in your site.", Path: r.URL.EscapedPath(), Method: r.Method,
		Hint: "Check the error below and the dev terminal. Fix the template or configuration, then save. Flea will validate the site again automatically."}
	if r.Method == "HEAD" {
		v.Method = "GET"
	}
	if cause != nil {
		v.Detail = cause.Error()
		if len(v.Detail) > 8192 {
			v.Detail = v.Detail[:8192] + " …"
		}
		var source *SourceError
		if errors.As(cause, &source) {
			v.File = source.File
			v.Line = source.Line
			v.Lines = source.Lines
		}
		switch {
		case strings.Contains(v.Detail, "unclosed if block") || strings.Contains(v.Detail, "unclosed unless block"):
			v.Hint = "Close this conditional with its matching {/if} or {/unless}. The highlighted line shows where it begins."
		case strings.Contains(v.Detail, "invalid JSON"):
			v.Hint = "Check the JSON file for missing commas, trailing commas and unclosed quotes. JSON keys and strings need double quotes. Save when corrected."
		case strings.Contains(v.Detail, "missing partial"):
			v.Hint = "Check the partial’s name and location. {render footer} looks for _footer.html beside the calling template; {render /footer} starts at the site root."
		case strings.Contains(v.Detail, "layout"):
			v.Hint = "Check the leading flea settings comment and the selected layout file. Layout paths start at the site root, and a layout needs {slot} for the page content."
		case strings.Contains(v.Detail, "route collision"):
			v.Hint = "Two HTML files map to the same URL. Keep just one: about.html and about/index.html both become /about."
		case strings.Contains(v.Detail, "snip") && (strings.Contains(v.Detail, "nested") || strings.Contains(v.Detail, "closing") || strings.Contains(v.Detail, "unclosed")):
			v.Hint = "Editable snip regions cannot be nested. Give each region balanced opening and closing tags, then save the HTML file."
		case strings.Contains(v.Detail, "expression") || strings.Contains(v.Detail, "directive"):
			v.Hint = "Check the template expression. Flea uses dotted names such as {app.site_name}, {render footer} for partials, and balanced {#if} or {#each} blocks."
		}
	}
	if code == 405 {
		v.Title = "This address doesn’t accept that request."
		v.Summary = "Website pages support GET and HEAD. Form submissions need a dedicated handler."
		v.Detail = "Method not allowed: " + r.Method
		v.Hint = "Check the form’s action and method. Contact forms submit to /_flea/contact; ordinary HTML pages cannot receive POST requests."
	}
	var out bytes.Buffer
	if err := devErrorTemplate.Execute(&out, v); err != nil {
		http.Error(w, "Development error page unavailable. Check the terminal.", code)
		return
	}
	script := reloadScript(revision)
	// The only executable code is the existing fixed reload watcher.
	scriptBody := script[len("<script>") : len(script)-len("</script>")]
	digest := sha256.Sum256(scriptBody)
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; script-src 'sha256-"+base64.StdEncoding.EncodeToString(digest[:])+"'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Frame-Options", "DENY")
	writeHTML(w, r, code, withReload(out.Bytes(), revision))
}
