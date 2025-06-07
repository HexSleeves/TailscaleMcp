import "dotenv/config";
import { TailscaleMCPServer, ServerMode } from "./server.js";
import { logger } from "./logger.js";

export interface CLIOptions {
  mode: ServerMode;
  port: number;
  help: boolean;
}

export class TailscaleMCPCLI {
  private options: CLIOptions;
  private server?: TailscaleMCPServer;

  constructor() {
    this.options = this.parseArguments();
  }

  private parsePort(args: string[]): number | undefined {
    const portIndex = args.indexOf("--port");
    if (portIndex !== -1 && args[portIndex + 1]) {
      const parsedPort = Number.parseInt(args[portIndex + 1], 10);
      if (!isNaN(parsedPort) && parsedPort >= 1 && parsedPort <= 65535) {
        return parsedPort;
      } else {
        throw new Error(
          `Invalid port number: ${
            args[portIndex + 1]
          }. Port must be between 1 and 65535.`
        );
      }
    }
    return undefined;
  }

  private parseArguments(): CLIOptions {
    const args = process.argv.slice(2);
    let mode: ServerMode = "stdio";
    let port = 3000;
    let help = false;

    // Check for --help flag
    if (args.includes("--help") || args.includes("-h")) {
      help = true;
    }

    // Check for --http flag
    if (args.includes("--http")) {
      mode = "http";
      port = this.parsePort(args) ?? 3000;
    }

    return { mode, port, help };
  }

  private showHelp(): void {
    console.log(`
Tailscale MCP Server

Usage:
  tailscale-mcp-server [options]

Options:
  --http              Start in HTTP mode (default: stdio mode)
  --port <number>     Port for HTTP mode (default: 3000)
  --help, -h          Show this help message

Examples:
  tailscale-mcp-server                    # Start in stdio mode (for MCP clients)
  tailscale-mcp-server --http             # Start HTTP server on port 3000
  tailscale-mcp-server --http --port 8080 # Start HTTP server on port 8080

Modes:
  stdio               Standard input/output mode for MCP clients like Claude Desktop
  http                HTTP server mode for testing and development with REST endpoints

Environment Variables:
  TAILSCALE_API_KEY   Your Tailscale API key (required for API operations)
  TAILSCALE_TAILNET   Your Tailscale tailnet name (required for API operations)
  LOG_LEVEL           Logging level: 0=DEBUG, 1=INFO, 2=WARN, 3=ERROR (default: 1)
  MCP_SERVER_LOG_FILE Log file path with optional {timestamp} placeholder
  NODE_ENV            Environment: development or production (affects security settings)
  CORS_ORIGIN         Allowed CORS origin for HTTP mode (production only)
    `);
  }

  async run(): Promise<void> {
    try {
      if (this.options.help) {
        this.showHelp();
        process.exit(0);
      }

      this.setupSignalHandlers();

      logger.debug("Starting Tailscale MCP Server...");
      this.server = new TailscaleMCPServer();
      await this.server.start(this.options.mode, this.options.port);
    } catch (error) {
      logger.error("Failed to start server:", error);
      await this.gracefulShutdown(1);
    }
  }

  private async gracefulShutdown(exitCode: number = 0): Promise<void> {
    try {
      await this.server?.stop();

      logger.debug("Flushing logs before shutdown...");
      await logger.flush();
      await logger.close();
    } catch (error) {
      console.error("Error during graceful shutdown:", error);
    } finally {
      process.exit(exitCode);
    }
  }

  setupSignalHandlers(): void {
    process.on("unhandledRejection", async (err) => {
      logger.error("Unhandled rejection:", err);
      await this.gracefulShutdown(1);
    });

    // Handle graceful shutdown
    process.on("SIGINT", () => {
      logger.debug("Received SIGINT, shutting down gracefully...");
      this.gracefulShutdown(0).catch((error) => {
        console.error("Error during SIGINT shutdown:", error);
        process.exit(1);
      });
    });

    process.on("SIGTERM", () => {
      logger.debug("Received SIGTERM, shutting down gracefully...");
      this.gracefulShutdown(0).catch((error) => {
        console.error("Error during SIGTERM shutdown:", error);
        process.exit(1);
      });
    });
  }
}
