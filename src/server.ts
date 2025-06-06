import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";

import { createTailscaleAPI, TailscaleCLI } from "./tailscale/index.js";
import { ToolRegistry } from "./tools/index.js";
import { logger } from "./logger.js";

export class TailscaleMCPServer {
  private server: Server;
  private toolRegistry!: ToolRegistry;

  constructor() {
    this.server = new Server(
      {
        name: "tailscale-mcp-server",
        version: "1.0.0",
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.setupHandlers();
  }

  async initialize(): Promise<void> {
    // Initialize Tailscale integrations
    const api = createTailscaleAPI();
    const cli = new TailscaleCLI();

    // Create tool registry and register tool modules
    this.toolRegistry = new ToolRegistry({ api, cli });
    await this.toolRegistry.loadTools();
  }

  private setupHandlers(): void {
    // List available tools
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: this.toolRegistry.getTools(),
      };
    });

    // Handle tool calls
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;
      return await this.toolRegistry.callTool(name, args || {});
    });
  }

  async start(): Promise<void> {
    await this.initialize();

    // Log server configuration
    logger.info("Tailscale MCP Server starting...");
    if (process.env.MCP_SERVER_LOG_FILE) {
      logger.info("File logging enabled");
    } else {
      logger.debug("File logging disabled (set MCP_SERVER_LOG_FILE to enable)");
    }

    const transport = new StdioServerTransport();
    await this.server.connect(transport);

    // Keep the process alive - different strategies for different environments
    if (process.stdin.isTTY) {
      // Interactive terminal - use stdin
      logger.debug("Interactive terminal detected, using stdin for keepalive");
      process.stdin.resume();
      process.stdin.setEncoding("utf8");

      process.stdin.on("end", () => {
        logger.info("stdin closed, shutting down MCP server");
        process.exit(0);
      });

      process.stdin.on("error", (error) => {
        logger.error("stdin error:", error);
        process.exit(1);
      });
    } else {
      // Non-interactive (Docker) - use interval keepalive
      logger.debug(
        "Non-interactive environment detected, using interval keepalive"
      );
      const keepAliveInterval = setInterval(() => {
        logger.debug("MCP Server keepalive heartbeat");
      }, 30000); // 30 second heartbeat

      // Store interval reference for cleanup
      process.on("exit", () => {
        clearInterval(keepAliveInterval);
      });
    }

    // Handle process termination gracefully
    const cleanup = () => {
      logger.info("MCP Server shutting down...");
      if (process.stdin.isTTY) {
        process.stdin.pause();
      }
      process.exit(0);
    };

    process.on("SIGINT", cleanup);
    process.on("SIGTERM", cleanup);

    logger.info(
      "Tailscale MCP Server started successfully and listening for MCP messages"
    );
  }
}

// For backwards compatibility and easy testing
export async function startServer(): Promise<void> {
  const server = new TailscaleMCPServer();
  await server.start();
}
