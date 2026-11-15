// Package serverunit holds the systemd units shared by setup and server updates.
package serverunit

import _ "embed"

//go:embed flea.service
var Service string

//go:embed flea.socket
var Socket string

// DirectService is the exact formerly installed unit. Only this unmodified unit
// may be converted automatically; it is also used to undo a failed conversion.
//
//go:embed direct.service
var DirectService string
