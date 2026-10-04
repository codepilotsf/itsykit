# One real Contact email

Opt-in operator check, not a unit test. Obtain approval for the exact sender,
recipient and destination host before uploading or running it. It sends one
clearly labeled message through a real managed Contact child and Flea's private
email capability, using the existing Resend settings in its environment.

```text
email-check -send -from approved@your-domain.com -to approved-recipient@example.com
```

Without all three explicit flags it exits without sending. The subject remains
Contact's fixed `New website contact message`; the body identifies the delivery
test. Reply-To uses the approved recipient. The tool checks its exact recipient,
sender override, Reply-To and rendered HTML/text before calling the real sender.
Only invalid submissions are retried for startup readiness. The one valid
submission is never retried, even after an uncertain timeout.

Compile `./scripts/contact-email-check` for the target platform after refreshing
the matching Contact bundle. Run as an unprivileged transient systemd unit with
PrivateTmp, ProtectSystem=strict, ProtectHome and NoNewPrivileges. Bind only the
exact uploaded directory read-only. Load the existing root-owned email settings
through systemd's EnvironmentFile; never print, copy or put credentials in argv.
The Contact child receives no provider credentials. Network access is required
for Resend HTTPS; the test's own TLS listener binds only to loopback.

The tool creates and removes its own temporary site and trusts only its local
test TLS certificate. It does not weaken production TLS, change the server
default sender, read live site configs or restart the live service. A 60-second
unit timeout, 192 MiB memory cap, no swap, 50% CPU quota and 64-task cap were
sufficient in the first Linux run. Verify the unit exits before removing its
exact uploaded directory.

Success checks provider acceptance, HTTP 303 and the thank-you flash. It does
not prove inbox receipt, spam placement, mailbox rendering or Reply-To behavior
in a real email client. Confirm those with the recipient before marking them done.
