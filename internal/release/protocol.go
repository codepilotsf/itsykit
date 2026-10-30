package release

// Protocol identifies the archive, host configuration and administration
// contract, independently of the CLI release version.
//
// 28 adds validated site-level migration redirects. Older servers silently
// ignore this config; matching clients and servers are required for deployment.
const Protocol = "FLEA_DEPLOY=28"
