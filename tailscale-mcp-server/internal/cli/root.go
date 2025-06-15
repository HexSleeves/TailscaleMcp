// tailscale-mcp-server/internal/cli/root.go
package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
	"github.com/hexsleeves/tailscale-mcp-server/version"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "tailscale-mcp-server",
	Short:   "Tailscale MCP Server",
	Version: version.Short(),
	Long: `A Model Context Protocol server that provides seamless integration with Tailscale's CLI commands and REST API, enabling automated network management and monitoring through a standardized interface.

Environment Variables:
  TAILSCALE_API_KEY        Tailscale API key (required for API operations)
  TAILSCALE_TAILNET        Tailnet name (required for API operations)
  TAILSCALE_API_BASE_URL   Custom API base URL (optional)
  LOG_LEVEL                Logging level: 0=debug, 1=info, 2=warn, 3=error (default: 1)
  MCP_SERVER_LOG_FILE      Log file path (optional)`,
	// Default behavior: show help if no subcommand
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			logger.Error("failed to display help", "error", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Custom version template
	rootCmd.SetVersionTemplate(version.Info() + "\n")
}

func initConfig() {
	// Logger level is handled during server initialization
}
