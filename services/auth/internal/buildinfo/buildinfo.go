// Package buildinfo exposes version metadata populated at build time via
// -ldflags "-X".
package buildinfo

// Version, Commit and Date are populated at build time via -ldflags "-X".
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
