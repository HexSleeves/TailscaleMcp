import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import express from "express";
import { randomUUID, createHash, randomBytes } from "node:crypto";
import * as http from "node:http";
import { logger } from "../logger.js";
import { ToolRegistry } from "../tools/index.js";

const TIMEOUT_MS = 15 * 60 * 1000; // 15 minutes

interface SessionInfo {
  transport: StreamableHTTPServerTransport;
  authToken: string;
  createdAt: Date;
  lastAccessed: Date;
  clientInfo?: {
    userAgent?: string;
    ip?: string;
  };
}

export class HttpMCPServer {
  private server: Server;
  private toolRegistry: ToolRegistry;
  private httpServer?: http.Server;
  private sessions: { [sessionId: string]: SessionInfo } = {};
  private cleanupInterval?: NodeJS.Timeout;

  constructor(toolRegistry: ToolRegistry) {
    this.toolRegistry = toolRegistry;
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

  private generateAuthToken(): string {
    return createHash("sha256")
      .update(randomUUID() + randomBytes(32).toString("hex"))
      .digest("hex");
  }

  private validateSession(
    sessionId: string,
    authToken: string,
    clientIp?: string
  ): SessionInfo | null {
    const session = this.sessions[sessionId];
    if (!session) {
      logger.warn(`Session validation failed: Session ${sessionId} not found`);
      return null;
    }

    if (session.authToken !== authToken) {
      logger.warn(
        `Session validation failed: Invalid auth token for session ${sessionId} from IP ${clientIp}`
      );
      return null;
    }

    // Update last accessed time
    session.lastAccessed = new Date();

    // Optional: Validate client IP consistency (can be disabled for development)
    if (
      process.env.NODE_ENV === "production" &&
      session.clientInfo?.ip &&
      clientIp &&
      session.clientInfo.ip !== clientIp
    ) {
      logger.warn(
        `Session validation failed: IP mismatch for session ${sessionId}. Expected ${session.clientInfo.ip}, got ${clientIp}`
      );
      return null;
    }

    return session;
  }

  private cleanupExpiredSessions(): void {
    const now = new Date();
    const expiredThreshold = 60 * 60 * 1000; // 1 hour

    Object.keys(this.sessions).forEach((sessionId) => {
      const session = this.sessions[sessionId];
      if (now.getTime() - session.lastAccessed.getTime() > expiredThreshold) {
        logger.debug(`Cleaning up expired session: ${sessionId}`);
        delete this.sessions[sessionId];
      }
    });
  }

  private setupRoutes(app: express.Application): void {
    // Add a simple health check endpoint
    app.get("/health", (req, res) => {
      res.json({
        status: "healthy",
        server: "tailscale-mcp-server",
        mode: "http",
        tools: this.toolRegistry.getTools().length,
      });
    });

    // Add a tools listing endpoint
    app.get("/tools", (req, res) => {
      res.json({ tools: this.toolRegistry.getTools() });
    });

    // Add session info endpoint for debugging (development only)
    app.get("/sessions", (req, res) => {
      if (process.env.NODE_ENV === "production") {
        res.status(404).json({ error: "Not found" });
        return;
      }

      const sessionSummary = Object.keys(this.sessions).map((sessionId) => ({
        sessionId,
        createdAt: this.sessions[sessionId].createdAt,
        lastAccessed: this.sessions[sessionId].lastAccessed,
        clientInfo: this.sessions[sessionId].clientInfo,
      }));

      res.json({
        activeSessions: sessionSummary.length,
        sessions: sessionSummary,
      });
    });

    // MCP POST endpoint handler with secure session management
    const mcpPostHandler = async (
      req: express.Request,
      res: express.Response
    ) => {
      try {
        const sessionId = req.headers["mcp-session-id"] as string;
        const authToken = req.headers["authorization"]?.replace(
          "Bearer ",
          ""
        ) as string;
        const clientIp = req.ip || req.socket.remoteAddress;
        const userAgent = req.headers["user-agent"];

        let sessionInfo: SessionInfo;

        if (sessionId && authToken) {
          // Validate existing session
          const validatedSession = this.validateSession(
            sessionId,
            authToken,
            clientIp
          );
          if (!validatedSession) {
            res.status(401).json({
              error: "Invalid session or authentication token",
              code: "INVALID_SESSION",
            });
            return;
          }
          sessionInfo = validatedSession;
        } else if (!sessionId && !authToken) {
          // Create new session with authentication
          const newAuthToken = this.generateAuthToken();
          const transport = new StreamableHTTPServerTransport({
            sessionIdGenerator: () => randomUUID(),
            onsessioninitialized: (newSessionId: string) => {
              // Changed from info to debug
              logger.debug(
                `New authenticated session initialized: ${newSessionId} from IP ${clientIp}`
              );

              sessionInfo = {
                transport,
                authToken: newAuthToken,
                createdAt: new Date(),
                lastAccessed: new Date(),
                clientInfo: {
                  userAgent,
                  ip: clientIp,
                },
              };

              this.sessions[newSessionId] = sessionInfo;

              // Send authentication token to client
              res.setHeader("X-Auth-Token", newAuthToken);
              res.setHeader("X-Session-ID", newSessionId);
            },
          });

          // Set up onclose handler to clean up session when transport closes
          transport.onclose = () => {
            const sid = transport.sessionId;
            if (sid && this.sessions[sid]) {
              logger.debug(
                `Transport closed for session ${sid}, removing from sessions`
              );
              delete this.sessions[sid];
            }
          };

          // Connect the server to this transport
          await this.server.connect(transport);

          // Store session info temporarily (will be updated in onsessioninitialized)
          sessionInfo = {
            transport,
            authToken: newAuthToken,
            createdAt: new Date(),
            lastAccessed: new Date(),
            clientInfo: { userAgent, ip: clientIp },
          };
        } else {
          // Missing either session ID or auth token
          res.status(400).json({
            error:
              "Both session ID and authorization token are required for existing sessions",
            code: "MISSING_CREDENTIALS",
          });
          return;
        }

        // Handle the request with the validated session
        await sessionInfo.transport.handleRequest(req, res);
      } catch (error) {
        logger.error("Error handling MCP request:", error);
        res.status(500).json({
          error: "Internal server error",
          code: "INTERNAL_ERROR",
        });
      }
    };

    // MCP GET endpoint handler (for SSE streams) with secure session validation
    const mcpGetHandler = async (
      req: express.Request,
      res: express.Response
    ) => {
      try {
        const sessionId = req.headers["mcp-session-id"] as string;
        const authToken = req.headers["authorization"]?.replace(
          "Bearer ",
          ""
        ) as string;
        const clientIp = req.ip || req.socket.remoteAddress;

        if (!sessionId || !authToken) {
          res.status(400).json({
            error: "Session ID and authorization token are required",
            code: "MISSING_CREDENTIALS",
          });
          return;
        }

        // Validate session ownership
        const sessionInfo = this.validateSession(
          sessionId,
          authToken,
          clientIp
        );
        if (!sessionInfo) {
          res.status(401).json({
            error: "Invalid session or authentication token",
            code: "INVALID_SESSION",
          });
          return;
        }

        // Handle the SSE request with the validated session
        await sessionInfo.transport.handleRequest(req, res);
      } catch (error) {
        logger.error("Error handling MCP GET request:", error);
        res.status(500).json({
          error: "Internal server error",
          code: "INTERNAL_ERROR",
        });
      }
    };

    // Set up MCP endpoints
    app.post("/mcp", mcpPostHandler);
    app.get("/mcp", mcpGetHandler);
  }

  async start(port: number): Promise<void> {
    logger.debug("Starting HTTP MCP server...");

    const app = express();

    // Enable JSON parsing
    app.use(express.json());

    // Add CORS headers for testing
    app.use((req, res, next) => {
      const allowedOrigin =
        process.env.NODE_ENV === "production"
          ? process.env.CORS_ORIGIN || "http://localhost:3000"
          : "*";

      res.header("Access-Control-Allow-Origin", allowedOrigin);
      res.header("Access-Control-Allow-Credentials", "true");
      res.header(
        "Access-Control-Allow-Methods",
        "GET, POST, PUT, DELETE, OPTIONS"
      );
      res.header(
        "Access-Control-Allow-Headers",
        "Origin, X-Requested-With, Content-Type, Accept, Authorization, mcp-session-id"
      );
      if (req.method === "OPTIONS") {
        res.sendStatus(200);
      } else {
        next();
      }
    });

    // Setup routes
    this.setupRoutes(app);

    // Run cleanup every 15 minutes
    this.cleanupInterval = setInterval(
      () => this.cleanupExpiredSessions(),
      TIMEOUT_MS
    );

    // Start the HTTP server
    this.httpServer = app.listen(port, () => {
      logger.debug(`HTTP server listening on port ${port}`);
      logger.debug(`Health check: http://localhost:${port}/health`);
      logger.debug(`Tools list: http://localhost:${port}/tools`);
      logger.debug(`MCP endpoint: http://localhost:${port}/mcp`);
    });

    // Handle process termination gracefully
    const cleanup = async () => {
      logger.debug("HTTP MCP Server shutting down...");
      await this.stop();
      process.exit(0);
    };

    process.on("SIGINT", cleanup);
    process.on("SIGTERM", cleanup);

    logger.debug(
      "HTTP MCP Server started successfully and listening for MCP messages"
    );
  }

  async stop(): Promise<void> {
    // Clear the session cleanup interval
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
      this.cleanupInterval = undefined;
    }

    // Clean up all active sessions
    Object.keys(this.sessions).forEach((sessionId) => {
      logger.debug(`Cleaning up session on shutdown: ${sessionId}`);
      delete this.sessions[sessionId];
    });

    if (this.httpServer) {
      await new Promise<void>((resolve) =>
        this.httpServer!.close(() => resolve())
      );
      this.httpServer = undefined;
    }

    logger.debug("HTTP MCP Server stopped");
  }
}
