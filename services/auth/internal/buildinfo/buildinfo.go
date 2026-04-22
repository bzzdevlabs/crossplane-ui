// Package buildinfo exposes version metadata populated at build time via
// -ldflags "-X".
package buildinfo

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
