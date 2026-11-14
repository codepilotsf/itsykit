package release

// Protocol identifies the archive, host configuration and administration
// contract, independently of the CLI release version.
//
// 29 adds the ItsyWeb site filenames, retaining explicit legacy Flea loading.
// Older servers cannot load new sites; matching clients/servers are required.
const Protocol = "FLEA_DEPLOY=29"
