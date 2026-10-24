# Hostname indexing policy

Flea can send `X-Robots-Tag: noindex, nofollow` for an explicitly configured
hostname while serving the same website release on another hostname without that
restriction. Website templates, app globals and content stay identical. This is
cooperative search indexing control, not authentication or a private preview.

This feature is unreleased and requires matching protocol-27 clients and servers.
The public 0.17.0 release uses protocol 26 and does not implement it. Ordinary
website deployment never installs a server executable. Upgrade and release work
require separate authorization and the existing release checks.

## Configure and inspect

After the matching client/server is installed, configure an already registered
managed hostname over SSH, from any directory:

```sh
flea server indexing root@server staging.example.com noindex
flea server list root@server
```

The command shows the current site and its aliases, asks you to type the exact
hostname, then changes only that hostname. Omit the SSH target to be prompted for
the server login. Non-root logins use the existing sudo flow. Updating policy
restarts the shared service, briefly interrupting other sites. It checks the
reviewed site state again under the deployment lock, validates the candidate,
journals the previous config, restarts and verifies a fresh readiness revision.
Failure restores the previous config; interrupted changes use
`sudo flea-server -recover` on the server. Pending deployment or executable
recovery blocks the change.

To stop adding the framework restriction to one hostname:

```sh
flea server indexing root@server staging.example.com allow
```

`allow` does not emit `index, follow` or remove owner-authored headers or robots
meta tags. It does not guarantee indexability. There is no automatic production
or staging classification. Use an exact lowercase registered hostname without a
port, trailing dot or wildcard. Unregistered hosts and other modes are rejected.
Configure each alias separately, including a WWW redirect source if needed.

`flea server list` shows the effective policy for every hostname, including the
default `allow`. `flea status` shows the remote site's effective policies after
checking its saved branch destination. These commands are read-only.

## Server configuration

The optional `indexing` map belongs to `/etc/flea/server.json`, alongside `sites`:

```json
{
  "listen": ":80",
  "sites": {
    "staging.example.com": "/var/lib/flea/sites/example/releases/RELEASE",
    "example.com": "/var/lib/flea/sites/example/releases/RELEASE"
  },
  "indexing": {
    "staging.example.com": "noindex",
    "example.com": "allow"
  }
}
```

This illustrates the schema; use the supported SSH command on installed servers
for locking, validation and recovery. Every policy key must be a registered host.
Existing configs without `indexing` load with unchanged behavior. The policy is
absent from `.flea.config.json`, portable releases and site backups. Branch
mapping settings do not set or reset it. Protocol 27 prevents an older client
from silently dropping the map through a typed config rewrite. There are no
legacy aliases, format migrations or an arbitrary response-header configuration.

Requests use Flea's existing Host normalization and registration check. Forwarded
headers and URL parameters cannot choose another policy. The outer host response
writer enforces the header at writes and flushes without buffering the body.
Aliases can share a cached site handler and keep separate policies. HTML, assets,
media, app responses, errors, managed downloads and HTTP/WWW redirects receive the
restriction, including GET and HEAD. Readiness at `/_flea/health` retains its
existing body, headers, method handling and unknown-host behavior. Existing TLS
allowlist and Host/SNI rejection checks stay in place.

## Crawling and preservation

Keep a public noindex site crawlable so crawlers can discover the header. Flea
does not rewrite robots.txt, sitemaps, canonicals or templates, and never generates
`noindex` or `nofollow` robots.txt directives or a blanket crawl block. A static
robots.txt is served unchanged on each alias. A site-wide `Disallow` rule can
prevent the noindex header from being discovered. Startup logs and server/status
inspection warn when the validated static robots.txt contains `Disallow: /`,
`/*` or `/*$`. This is a conservative diagnostic: review user-agent groups and
Allow exceptions yourself. Files above the in-memory asset limit produce a
manual-review warning rather than being buffered for analysis. This diagnostic
does not interpret dynamic robots routes or every crawler's rule syntax.

Normal deployment, deploying another site, explicit content sync, code rollback
and site-backup restore retain the **current** host policies. A site backup does
not capture or restore policy. Code rollback does not roll policy back in time.
A failed or interrupted host-config transaction restores its complete prior
configuration, including the prior policies, through the existing recovery
journal. This is distinct from code rollback and site restore.

Removing a site unregisters its aliases and deletes only those aliases' indexing
entries; unrelated policies remain. Its releases and other retained files stay
on disk. Redeploying a removed name does not restore its old host policies or
retained content. Record policies separately and explicitly configure them again.
Whole-server disaster recovery still requires a separate protected server-config
backup. See [recovery](recovery.md).

## CodePilot rollout and rollback

These are future operator steps, not actions performed by the implementation:

1. Review the change. Complete release verification for a new minor release and
   install its matching client and server through the supported release and
   `flea update-server` process, with separate authorization. Keep the website's
   existing `app.prevent_indexing` and conditional robots meta protection.
2. Inspect `flea server list root@fleacms.com`. Configure only the already
   registered staging hostname:

   ```sh
   flea server indexing root@fleacms.com staging.codepilot.com noindex
   flea server list root@fleacms.com
   curl -sSI https://staging.codepilot.com/
   curl -sS -D - https://staging.codepilot.com/robots.txt
   ```

   Type `staging.codepilot.com` when asked. Verify the actual response has
   `X-Robots-Tag: noindex, nofollow`, robots.txt permits crawling, and public assets,
   a missing page and the HTTP redirect also carry the header. Check an existing
   unaffected hosted domain remains unchanged. Do not register `codepilot.com`,
   change DNS or alter another host's policy to demonstrate this feature.
3. Only after live verification, make a separate consumer change removing
   `app.prevent_indexing` and its conditional meta blocks. Keep robots.txt
   crawlable. Update the production guard and consumer documentation to check
   the native policy while preserving other launch checks. Preserve the signup
   URL `https://hub.codepilot.com/get-started`, Hub code and billing settings.
4. Deploy that consumer change only with its own authorization. Verify the live
   staging header again without the app flag or meta blocks. Record installed
   versions, policy, local test results and live results separately.

If rollout stops, retain the existing meta protection. If native protection must
be disabled on staging after consumer cleanup, first restore and verify the
consumer meta fallback, then explicitly set `allow` if intended. A website code
rollback alone keeps the native policy. A failed policy transaction restores its
prior policy; inspect it after `-recover`. Do not downgrade the server to protocol
26 while an `indexing` map is present: that server rejects the new config. Any
executable/config rollback needs its own reviewed server-maintenance procedure
and authorization, with staging protection verified throughout. No downgrade or
migration mechanism is added here.
