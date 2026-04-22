// Package buildinfo exposes version metadata populated at build time via
// -ldflags "-X". See Taskfile.yml for the values injected by the build.
package buildinfo

// These variables are overwritten at link time.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
