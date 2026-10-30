# Search discovery and canonical URLs

From 0.20.0, Flea can generate a sitemap, advertise it through robots.txt, add canonical URLs
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
This feature requires matching CLI/server deployment protocol 28; older servers
ignore SEO settings. Deploying site configuration alone does not upgrade the
server executable.

## Enable, check and deploy

1. Update your local CLI with `flea update`, then check `flea version`. This feature
   starts in 0.20.0. For an existing older server, run `flea update-server` from a
   site whose current Git branch has the intended server saved as its deployment
   destination. It confirms before briefly restarting every site on that host.
   A new server set up with the current CLI already receives its bundled runtime.
2. Merge the `seo` settings above into your existing configuration. Keep your
   app, admin, apps and deployment settings. Use the production address even when
   working locally or on staging. Omit `exclude` if no additional page needs it.
3. Run `flea status`, then `flea dev`. Open the address printed by dev and check
   `/sitemap.xml` and `/robots.txt`. The XML should contain your production URLs,
   including while served locally. View a public page's source to find its
   canonical tag. A page listed in `exclude` should still load but return a
   `X-Robots-Tag: noindex` response header.
4. Commit your reviewed configuration and site changes, then run `flea deploy`
   from the intended branch. Normal deployment publishes the committed
   configuration and preserves existing live editable content. Repeat the URL,
   canonical and exclusion checks on the production domain.

The example paths are placeholders: replace them with real routes on your site.
An excluded page that does not exist still returns 404; exclusion does not create
it. In a browser, use View Source for canonical markup and the Network panel for
response headers. If you use curl, an existing excluded page can be checked with:

```sh
curl -I https://example.com/campaign/receipt
```

Expect `X-Robots-Tag: noindex`. A fragment without a closing head uses a canonical
`Link` response header instead of an HTML tag. Custom discovery files take
precedence, as explained below; they can account for a sitemap or robots.txt that
does not match the automatic output.

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
Migration redirects cannot use `/sitemap.xml` or `/robots.txt` as a source while
automatic SEO is enabled. Old page redirects remain outside the sitemap and do
not receive automatic canonical markup.

[Host indexing policy](host-indexing.md) remains separate, in server configuration.
A staging hostname configured as noindex still receives its noindex/nofollow
header on pages, the sitemap and robots.txt. Its shared sitemap and canonical
URLs still name the configured production origin. SEO settings never remove that
host restriction or add a crawl block. Configure staging policy explicitly; a
production URL does not automatically classify other hosts as staging.

## Optional Google Search Console submission

Google Search Console is Google's free reporting tool for website owners. Flea
does not require an account or a Search Console integration to generate a sitemap.
Its generated robots.txt already advertises the sitemap so Google can discover it.

To check whether Google has read the sitemap and reported processing errors:

1. Open [Search Console](https://search.google.com/search-console) and verify that
   you own the production domain or site address.
2. Select that site and open **Sitemaps**.
3. Submit `https://example.com/sitemap.xml`. If the form already supplies the site
   address, enter only `sitemap.xml`. You are submitting the URL, not uploading XML.

Do this after the sitemap is reachable on your public domain. The same URL stays
current as public content changes; you do not need to submit it after every edit.
This step is optional and does not guarantee indexing, ranking or AI citations.
See [Google's sitemap guide](https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap).
