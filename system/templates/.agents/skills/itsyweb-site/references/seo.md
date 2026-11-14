# Search discovery and host indexing

Read the installed SEO guide before changing search behavior. Use the selected
config, preserving other settings. Optional top-level `seo`, alongside `app`,
enables automatic discovery:

```json
{
  "app": {"site_name": "Example"},
  "seo": {"url": "https://example.com", "exclude": ["/campaign/receipt"]}
}
```

`url` is the production HTTPS origin even in dev/staging, with no credentials,
subdirectory, query or fragment. `exclude` is optional and matches exact page
paths, without wildcards, queries, `.html` aliases or trailing slashes. Exclusion
leaves a page accessible, adding noindex and omitting it from automatic discovery.
It is not access control. Do not block crawlers from reading the noindex response.

Without `seo`, automatic discovery is disabled. With it, the immutable snapshot
supplies `/sitemap.xml`, default `/robots.txt` with a Sitemap line, and missing
page canonicals using only the configured production origin. It never derives
production URLs from request headers or writes generated files into the site.
Fragments without a closing head use an HTTP Link canonical header.

Custom robots/sitemap routes and authored canonical tags win. Maintain discovery
and exclusions yourself in custom files. Eligible ordinary pages, Contact forms,
Blog listings and published posts enter automatic discovery; private/admin/error
routes, assets, drafts, Contact receipts, custom managed apps and redirects do
not. Existing noindex and incompatible custom canonicals also affect eligibility.
Titles, descriptions and structured data remain the author's responsibility.

Server exact-host indexing policy is separate from site `seo` and `app` settings.
`itsy server indexing root@server staging.example.com noindex` changes a registered
hostname after confirmation and restarts the shared service. `allow` removes only
that imposed restriction. Policy survives deploy, sync, rollback and restore;
site backups exclude it. `itsy server list` and status report effective policy.
A production SEO origin does not automatically protect staging. Both systems
require a supporting server; new-layout deployment requires protocol 30 and an
explicit authorized server upgrade, not merely changed site config.

Verify locally: sitemap absolute production URLs, robots discovery, public
canonical tags/Link headers, excluded page accessibility/noindex, custom-file
precedence and unchanged content. Changing local files does not authorize host
policy changes, deployment or Search Console submission. Search Console is
optional; generating a sitemap does not guarantee indexing or ranking.
