package release

// Protocol identifies the archive/content contract, independently of the CLI
// release version. Change it when old servers cannot safely accept new sites.
const Protocol = "NANOWEB_DEPLOY=7"
