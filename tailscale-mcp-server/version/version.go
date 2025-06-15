// tailscale-mcp-server/version/version.go
package version

import (
	"fmt"
	"runtime"
)

var (
	// These will be set by the build process
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// Info returns formatted version information
func Info() string {
	return fmt.Sprintf("tailscale-mcp-server %s (commit: %s, built: %s, go: %s)",
		Version, GitCommit, BuildTime, GoVersion)
}

// Short returns just the version
func Short() string {
	return Version
}
