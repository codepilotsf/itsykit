package release

// Protocol identifies the archive, host configuration and administration
// contract, independently of the CLI release version.
//
// 27 adds exact-host indexing policy. Older typed config writers could discard
// that policy; matching clients and servers are required for all administration.
const Protocol = "FLEA_DEPLOY=27"
