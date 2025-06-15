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
	Platform  = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Info returns formatted version information
func Info() string {
	output := fmt.Sprintf("Tailscale MCP Server [%s]\n", Version)
	output += fmt.Sprintf("Built with [%s]\n", GoVersion)
	output += fmt.Sprintf("Platform: [%s]\n", Platform)

	return output
}

// Short returns just the version
func Short() string {
	return Version
}
