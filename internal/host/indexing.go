package host

import (
	"fmt"
	"net/http"
	"strings"

	"flea/internal/site"
)

// ValidIndexingMode accepts the two exact-host policies. Allow adds no header
// and never removes an owner's own restriction.
func ValidIndexingMode(mode string) bool { return mode == "allow" || mode == "noindex" }

func validateIndexing(c Config) error {
	for domain, mode := range c.Indexing {
		if !ValidHost(domain) || c.Sites[domain] == "" {
			return fmt.Errorf("indexing requires an exact registered hostname: %q", domain)
		}
		if !ValidIndexingMode(mode) {
			return fmt.Errorf("invalid indexing mode for %s: use noindex or allow", domain)
		}
	}
	return nil
}

// IndexingWarnings reports conservative static robots.txt diagnostics from the
// validated snapshot. It is not a complete crawler-specific robots evaluator.
func (c Config) IndexingWarnings(domain string) []string { return c.indexingWarnings[domain] }

func robotsCrawlWarning(p *site.Project) string {
	data, ok := p.AssetData["/robots.txt"]
	if !ok {
		if _, exists := p.Assets["/robots.txt"]; exists {
			return "robots.txt is too large for the crawl-rule diagnostic; review it to ensure crawlers can discover the noindex header."
		}
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line, _, _ = strings.Cut(line, "#")
		key, value, _ := strings.Cut(line, ":")
		value = strings.TrimSpace(value)
		if strings.EqualFold(strings.TrimSpace(key), "disallow") && (value == "/" || value == "/*" || value == "/*$") {
			return "robots.txt contains a site-wide Disallow rule; affected crawlers cannot discover the noindex header. Review its user-agent groups and allow crawling. Flea leaves the file unchanged."
		}
	}
	return ""
}

// Select by the same registered Host used for routing, never forwarding headers.
// Apply outside root-cached handlers so aliases can have independent policies.
// Health retains its existing public readiness contract.
func indexingHandler(c Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := requestHost(r)
		if r.URL.Path == HealthPath || c.Sites[domain] == "" || c.Indexing[domain] != "noindex" {
			next.ServeHTTP(w, r)
			return
		}
		restricted := &indexingWriter{ResponseWriter: w}
		restricted.restrict()
		next.ServeHTTP(restricted, r)
		// A handler can return without Write/WriteHeader, including on HEAD.
		restricted.restrict()
	})
}

// Enforce at header commitment and flush without buffering bodies. Unwrap keeps
// ResponseController deadlines and other supported transport operations working.
type indexingWriter struct{ http.ResponseWriter }

func (w *indexingWriter) restrict() {
	// Header maps can also contain noncanonical keys written directly by handlers.
	for key := range w.Header() {
		if strings.EqualFold(key, "X-Robots-Tag") {
			delete(w.Header(), key)
		}
	}
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
}
func (w *indexingWriter) WriteHeader(code int) {
	w.restrict()
	w.ResponseWriter.WriteHeader(code)
}
func (w *indexingWriter) Write(p []byte) (int, error) {
	w.restrict()
	return w.ResponseWriter.Write(p)
}
func (w *indexingWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *indexingWriter) FlushError() error {
	w.restrict()
	return http.NewResponseController(w.ResponseWriter).Flush()
}
func (w *indexingWriter) Flush() { _ = w.FlushError() }
