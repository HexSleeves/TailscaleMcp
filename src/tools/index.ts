import type { CallToolResult, Tool } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod/v4";
import { TailscaleAPI } from "../tailscale/tailscale-api.js";
import { TailscaleCLI } from "../tailscale/tailscale-cli.js";

// Import all tool modules
import { logger } from "../logger.js";
import { aclTools } from "./acl-tools.js";
import { adminTools } from "./admin-tools.js";
import { deviceTools } from "./device-tools.js";
import { networkTools } from "./network-tools.js";

export interface ToolContext {
  api: TailscaleAPI;
  cli: TailscaleCLI;
}

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: z.ZodSchema;
  handler: (args: any, context: ToolContext) => Promise<CallToolResult>;
}

export interface ToolModule {
  tools: ToolDefinition[];
}

// Tool registry
export class ToolRegistry {
  private tools = new Map<string, ToolDefinition>();
  private context: ToolContext;

  constructor(context: ToolContext) {
    this.context = context;
  }

  async loadTools(): Promise<void> {
    logger.debug("Loading tools...");

    // Register all tool modules
    this.registerModule(deviceTools);
    this.registerModule(networkTools);
    this.registerModule(aclTools);
    this.registerModule(adminTools);

    logger.debug(`Loaded ${this.tools.size} tools`);
  }

  register(tool: ToolDefinition): void {
    if (this.tools.has(tool.name)) {
      logger.warn(
        `Duplicate tool name detected: "${tool.name}" – overriding previous definition`
      );
    }
    this.tools.set(tool.name, tool);
  }

  registerModule(module: ToolModule): void {
    for (const tool of module.tools) {
      this.register(tool);
    }
  }

  getTools(): Tool[] {
    return Array.from(this.tools.values()).map((tool) => ({
      name: tool.name,
      description: tool.description,
      inputSchema: zodToJsonSchema(tool.inputSchema),
    }));
  }

  async callTool(name: string, args: any): Promise<CallToolResult> {
    const tool = this.tools.get(name);
    if (!tool) {
      return {
        content: [{ type: "text", text: `Unknown tool: ${name}` }],
        isError: true,
      };
    }

    try {
      const validatedArgs = tool.inputSchema.safeParse(args);
      if (!validatedArgs.success) {
        logger.warn("Invalid arguments:", validatedArgs.error);

        return {
          content: [
            {
              type: "text",
              text: `Invalid arguments: ${validatedArgs.error.message}`,
            },
          ],
          isError: true,
        };
      }

      return await tool.handler(validatedArgs.data, this.context);
    } catch (error: any) {
      logger.error("Tool error:", error);

      return {
        content: [{ type: "text", text: `Tool error: ${error.message}` }],
        isError: true,
      };
    }
  }

  /**
   * Dispose of resources held by the ToolRegistry and its context
   */
  async dispose(): Promise<void> {
    logger.debug("Disposing ToolRegistry resources...");

    // Clear the tools map
    this.tools.clear();

    // Note: TailscaleAPI uses Axios which doesn't require explicit cleanup
    // as it doesn't maintain persistent connections by default.
    // If in the future the API client maintains persistent connections,
    // we would add cleanup logic here.

    logger.debug("ToolRegistry resources disposed");
  }
}

// Helper function to convert Zod schema to JSON Schema
function zodToJsonSchema(schema: z.ZodSchema): any {
  try {
    return z.toJSONSchema(schema);
  } catch (error) {
    logger.error("Schema conversion failed:", error);
    return { type: "object", properties: {} };
  }
}
