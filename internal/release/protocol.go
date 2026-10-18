package release

// Protocol identifies the archive/content contract, independently of the CLI
// release version. Change it when old servers cannot safely accept new sites.
//
// 26 adds the custom-app package contract: a release may carry confined
// flea/apps/<name>/app.json manifests and per-platform native executables,
// extracted 0555 (readable and executable by the service user, writable by no
// one). A protocol-25 server cannot extract a package and is refused.
const Protocol = "FLEA_DEPLOY=26"
