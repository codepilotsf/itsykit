package site

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	xhtml "golang.org/x/net/html"
)

// SEO is opt-in. The production origin is site-owned configuration, never a
// request Host or forwarding header. Generated documents belong to the snapshot.
type seoSettings struct {
	URL     string   `json:"url"`
	Exclude []string `json:"exclude"`
}

func decodeSEO(raw json.RawMessage) (seoSettings, error) {
	var cfg seoSettings
	if raw == nil {
		return cfg, nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return cfg, fmt.Errorf("seo must be an object")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("seo must contain only url and exclude settings")
	}
	u, err := url.Parse(cfg.URL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil ||
		u.Opaque != "" || (u.Path != "" && u.Path != "/") || u.RawPath != "" ||
		u.RawQuery != "" || u.ForceQuery || u.Fragment != "" ||
		strings.ContainsAny(cfg.URL, "\\\t\r\n <>\"#") || !seoHost(u) {
		return cfg, fmt.Errorf("seo.url must be an HTTPS origin, such as https://example.com, without credentials, a path, query or fragment")
	}
	// A trailing slash is harmless; keep one representation for all outputs.
	u.Path = ""
	u.Host = strings.ToLower(u.Host)
	cfg.URL = u.String()
	seen := map[string]bool{}
	for _, route := range cfg.Exclude {
		if !safeURL(route) || strings.ContainsAny(route, "?#*\t\r\n ") || route == "/admin" || strings.HasPrefix(route, "/admin/") {
			return cfg, fmt.Errorf("seo.exclude entries must be exact public page paths, such as /contact/thank-you")
		}
		if seen[route] {
			return cfg, fmt.Errorf("seo.exclude contains a duplicate page path")
		}
		seen[route] = true
	}
	return cfg, nil
}

func seoHost(u *url.URL) bool {
	if strings.HasSuffix(u.Host, ":") {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.ParseUint(port, 10, 16)
		if err != nil || n == 0 {
			return false
		}
	}
	host := strings.ToLower(u.Hostname())
	if ip, err := netip.ParseAddr(host); err == nil {
		return ip.Zone() == ""
	}
	if len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}

// Managed apps can require sessions and do not yet declare public read routes.
// Their templates alone are insufficient evidence for sitemap discovery. Blog
// and Contact have known public routes; Contact's receipt is not a landing page.
func (p *Project) seoPublicPage(route string) bool {
	if p.Pages[route] == nil {
		return false
	}
	for name, app := range p.Apps {
		if route != app.Mount && !strings.HasPrefix(route, app.Mount+"/") {
			continue
		}
		return name == "blog" || (name == "contact" && route == app.Mount)
	}
	return true
}

func (p *Project) canonicalURL(route string) string {
	if p.seo.URL == "" || p.seoExcluded[route] || p.seoPages[route].noindex || !p.seoPublicPage(route) {
		return ""
	}
	return p.seo.URL + (&url.URL{Path: route}).EscapedPath()
}

func (p *Project) loadSEO() error {
	if p.seo.URL == "" {
		return nil
	}
	p.seoExcluded = map[string]bool{}
	for _, route := range p.seo.Exclude {
		p.seoExcluded[route] = true
	}
	if app, ok := p.Apps["contact"]; ok {
		p.seoExcluded[app.Mount+"/thank-you"] = true
	}
	// An authored page or asset at either endpoint takes precedence. Do not
	// overwrite, append to, buffer or reinterpret a custom discovery document.
	if p.Pages["/sitemap.xml"] == nil && p.Assets["/sitemap.xml"] == "" {
		var body bytes.Buffer
		body.WriteString(xml.Header + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
		count := 0
		for _, route := range p.Routes() {
			if route == "/robots.txt" || route == "/sitemap.xml" {
				continue
			}
			location := p.canonicalURL(route)
			if location == "" {
				continue
			}
			if meta := p.seoPages[route]; meta.hasCanonical {
				if meta.canonical != location {
					continue
				}
			}
			entry, err := xml.Marshal(struct {
				XMLName xml.Name `xml:"url"`
				Loc     string   `xml:"loc"`
			}{Loc: location})
			if err != nil {
				return fmt.Errorf("generate sitemap: %w", err)
			}
			count++
			if count > 50000 || body.Len()+len(entry)+len("\n</urlset>\n") > 50<<20 {
				return fmt.Errorf("automatic sitemap exceeds 50,000 URLs or 50 MiB; provide site/sitemap.xml with a sitemap index")
			}
			body.Write(entry)
			body.WriteByte('\n')
		}
		body.WriteString("</urlset>\n")
		p.seoSitemap = body.Bytes()
	}
	if p.Pages["/robots.txt"] == nil && p.Assets["/robots.txt"] == "" {
		p.seoRobots = []byte("User-agent: *\nDisallow:\n\nSitemap: " + p.seo.URL + "/sitemap.xml\n")
	}
	return nil
}

func (p *Project) serveSEO(w http.ResponseWriter, r *http.Request) bool {
	var body []byte
	contentType := ""
	switch r.URL.Path {
	case "/sitemap.xml":
		body, contentType = p.seoSitemap, "application/xml; charset=utf-8"
	case "/robots.txt":
		body, contentType = p.seoRobots, "text/plain; charset=utf-8"
	}
	if body == nil {
		return false
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprint(len(body)))
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
	return true
}

// Add only the missing canonical, without reserializing site-owned HTML. A
// custom canonical wins. Fragments without a head use the standard Link header.
func (p *Project) canonicalResponse(w http.ResponseWriter, route string, body []byte) []byte {
	canonical := p.canonicalURL(route)
	if canonical == "" || route == "/robots.txt" || route == "/sitemap.xml" {
		return body
	}
	meta := inspectSEO(body)
	if meta.hasCanonical || meta.noindex {
		return body
	}
	if meta.headEnd < 0 {
		w.Header().Add("Link", "<"+canonical+">; rel=\"canonical\"")
		return body
	}
	tag := []byte(`<link rel="canonical" href="` + html.EscapeString(canonical) + `">` + "\n")
	result := make([]byte, 0, len(body)+len(tag))
	result = append(result, body[:meta.headEnd]...)
	result = append(result, tag...)
	return append(result, body[meta.headEnd:]...)
}

type seoHTML struct {
	headEnd      int
	hasCanonical bool
	canonical    string
	noindex      bool
}

// Tokenize only the head. Comments, scripts and inert template contents cannot
// masquerade as metadata. Record offsets to preserve every original source byte.
func inspectSEO(body []byte) seoHTML {
	meta := seoHTML{headEnd: -1}
	z := xhtml.NewTokenizer(bytes.NewReader(body))
	offset, templateDepth, inHead := 0, 0, true
	for {
		kind := z.Next()
		start := offset
		offset += len(z.Raw())
		if kind == xhtml.ErrorToken {
			break
		}
		if kind != xhtml.StartTagToken && kind != xhtml.SelfClosingTagToken && kind != xhtml.EndTagToken {
			continue
		}
		token := z.Token()
		if token.Data == "template" {
			if kind == xhtml.EndTagToken && templateDepth > 0 {
				templateDepth--
			} else if kind == xhtml.StartTagToken {
				templateDepth++
			}
			continue
		}
		if templateDepth > 0 {
			continue
		}
		if token.Data == "head" {
			if kind == xhtml.StartTagToken {
				inHead = true
			} else if kind == xhtml.EndTagToken && inHead {
				meta.headEnd = start
				break
			}
		}
		if kind == xhtml.StartTagToken || kind == xhtml.SelfClosingTagToken {
			switch token.Data {
			case "html", "head", "base", "basefont", "bgsound", "link", "meta", "title", "noscript", "noframes", "script", "style":
			default:
				return meta // The body may start without an explicit <body>.
			}
		}
		if inHead && token.Data == "link" && kind != xhtml.EndTagToken {
			canonical, href := false, ""
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					href = attr.Val
				}
				if attr.Key == "rel" {
					for _, rel := range strings.Fields(attr.Val) {
						if strings.EqualFold(rel, "canonical") {
							canonical = true
						}
					}
				}
			}
			if canonical {
				if meta.hasCanonical && meta.canonical != href {
					href = "" // Conflicting canonicals cannot enter the sitemap.
				}
				meta.hasCanonical, meta.canonical = true, href
			}
		}
		if inHead && token.Data == "meta" && kind != xhtml.EndTagToken {
			name, content := "", ""
			for _, attr := range token.Attr {
				if attr.Key == "name" {
					name = strings.ToLower(attr.Val)
				} else if attr.Key == "content" {
					content = strings.ToLower(attr.Val)
				}
			}
			if name == "robots" || name == "googlebot" {
				for _, directive := range strings.FieldsFunc(content, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\r' || r == '\n' }) {
					if directive == "noindex" || directive == "none" {
						meta.noindex = true
					}
				}
			}
		}
	}
	return meta
}
