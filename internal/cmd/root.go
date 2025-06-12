package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/hexsleeves/tailscale-mcp-server/internal/logger"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tailscale-mcp-server",
	Short: "Tailscale MCP Server",
	Long: `A Model Context Protocol server that provides seamless integration
with Tailscale's CLI commands and REST API, enabling automated network
management and monitoring through a standardized interface.

Environment Variables:
  TAILSCALE_API_KEY        Tailscale API key (required for API operations)
  TAILSCALE_TAILNET        Tailnet name (required for API operations)
  TAILSCALE_API_BASE_URL   Custom API base URL (optional)
  LOG_LEVEL                Logging level: 0=debug, 1=info, 2=warn, 3=error (default: 1)
  MCP_SERVER_LOG_FILE      Log file path (optional)`,
	Version: getVersion(),
	// Default behavior: show help if no subcommand
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Logger level is handled during server initialization
}

func getVersion() string {
	// This will be replaced during build with actual version via ldflags
	return "dev"
}

func init() {
	// Custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("Tailscale MCP Server %s\nBuilt with %s\nPlatform: %s/%s\n",
		getVersion(),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH))
}
