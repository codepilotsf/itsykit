# Search discovery and canonical URLs

Flea can generate a sitemap, advertise it through robots.txt, add canonical URLs
and exclude selected pages from search. Enable this in `.flea.config.json`,
preserving the other settings:

```json
{
  "app": {"site_name": "Example"},
  "seo": {
    "url": "https://example.com",
    "exclude": ["/contact/thank-you", "/campaign/receipt"]
  }
}
```

`seo.url` is the production HTTPS origin. It must not contain credentials, a
subdirectory, query or fragment. An optional trailing slash is normalized away;
hostnames are normalized to lowercase. Use an ASCII hostname (Punycode for an
internationalized domain). Numeric ports from 1 to 65535 are accepted. This setting
does not register hosts, redirect aliases, configure TLS or deploy a site. Flea
never selects this address from a request Host or forwarding header.

Without `seo`, automatic SEO behavior is disabled. Adding the object requires
`url`; `exclude` is optional. Unknown SEO settings and malformed values fail site
validation. No additional first-run questions or external services are involved.
This feature requires a server built with the SEO implementation; deploying site
configuration alone does not upgrade the server executable.

## Sitemap

`GET /sitemap.xml` returns UTF-8 XML with absolute production URLs in stable
route order. It includes ordinary public HTML pages, the enabled Blog listing
and published posts, and the enabled Contact form. Admin pages, private files,
assets, errors, blog drafts and Contact's thank-you page are excluded. Custom
managed app routes are omitted because their templates do not declare whether
the app requires a session. Use a custom sitemap for public app routes.

Pages containing a `robots` or `googlebot` meta tag with `noindex` or `none` are
omitted. A custom canonical is preserved; the page enters the generated sitemap
only if its canonical exactly matches the absolute production URL Flea would
generate. Pages with relative, empty or different custom canonicals are omitted.
Metadata is checked in the ordinary visitor rendering at load time, without
issuing form tokens, starting apps or rendering authenticated content.

The sitemap is part of the immutable site snapshot. Publishing, unpublishing and
content saves replace it along with the public routes; source reloads rebuild it.
Generation writes no site files. HEAD returns the same headers without a body;
unsupported methods receive 405. Dev does not inject reload JavaScript into XML.

Flea omits `lastmod`, `priority` and `changefreq`. It does not track the last
significant change to a whole rendered page, so a deployment timestamp would be
misleading. A generated sitemap is limited to 50,000 URLs and 50 MiB. Beyond
either limit, validation asks for a custom sitemap index instead of publishing
an incomplete sitemap.

## Canonical URLs

On successful public page responses, Flea inserts a missing
`<link rel="canonical" href="https://example.com/page">` before the closing
`</head>`. Source files and existing markup bytes remain unchanged. HTML
fragments without a closing head receive an HTTP `Link` canonical header instead.
Query parameters are not part of the canonical URL.

An existing canonical tag takes precedence. Flea does not add a canonical to
excluded or noindex pages, errors, admin pages, custom managed apps, discovery
documents, redirects or unsuccessful app responses. Enabled Contact's form and
Blog's public pages use the same production origin as the sitemap. This does not
replace titles, descriptions, structured data or custom layouts.

## Excluding pages

Each `seo.exclude` entry is an exact public page path, such as `/receipt` or `/`.
No wildcards, queries, fragments, `.html` aliases or trailing slashes are accepted.
An entry affects that page only, not its descendants. Paths may be configured
before the page exists. Excluded pages remain accessible, receive
`X-Robots-Tag: noindex`, and are omitted from the generated sitemap and automatic
canonicals. The enabled Contact thank-you page receives this treatment by default.
Existing custom canonical markup is still preserved.

Exclusion is cooperative search control, not authentication. Do not put secrets
on a public page. Do not block an excluded page in robots.txt: crawlers need to
fetch it to discover the noindex response.

## Robots.txt and custom discovery files

When no custom robots.txt route exists, Flea generates:

```text
User-agent: *
Disallow:

Sitemap: https://example.com/sitemap.xml
```

An authored `site/robots.txt` or `site/sitemap.xml` takes precedence independently.
Flea serves those files unchanged, including streaming large files through the
existing pinned asset handles. It does not append a Sitemap line to custom
robots.txt or remove excluded URLs from a custom sitemap. Add the production
Sitemap line and maintain exclusions yourself when providing these files.
An authored HTML page mapped to either discovery endpoint also takes precedence.

[Host indexing policy](host-indexing.md) remains separate, in server configuration.
A staging hostname configured as noindex still receives its noindex/nofollow
header on pages, the sitemap and robots.txt. Its shared sitemap and canonical
URLs still name the configured production origin. SEO settings never remove that
host restriction or add a crawl block. Configure staging policy explicitly; a
production URL does not automatically classify other hosts as staging.

After publication, verify the production property in Search Console and submit
the sitemap. Discovery, indexing, ranking and AI citations are not guaranteed.
