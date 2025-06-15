package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/hexsleeves/tailscale-mcp-server/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display comprehensive version information including:
- Application version
- Go runtime version
- Build platform
- Compilation details

This is useful for debugging and support purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Tailscale MCP Server %s\n", version.Info())
		fmt.Printf("Built with %s\n", runtime.Version())
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		if verbose {
			// Additional verbose information
			fmt.Printf("Go max procs: %d\n", runtime.GOMAXPROCS(0))
			fmt.Printf("Go routines: %d\n", runtime.NumGoroutine())
			fmt.Printf("Go compiler: %s\n", runtime.Compiler)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
