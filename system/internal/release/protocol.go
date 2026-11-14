package release

// Protocol identifies the archive, host configuration and administration
// contract, independently of the CLI release version.
//
// 30 removes page CMS content and retains Blog posts and cover media.
// Older servers cannot load new sites; matching clients/servers are required.
const Protocol = "FLEA_DEPLOY=30"
