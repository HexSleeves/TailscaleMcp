import type { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod/v4";
import { logger } from "../logger.js";
import type { ToolContext, ToolModule } from "./index.js";

// Schemas
const NetworkStatusSchema = z.object({
  format: z
    .enum(["json", "summary"])
    .optional()
    .default("json")
    .describe("Output format (json or summary)"),
});

const ConnectNetworkSchema = z.object({
  acceptRoutes: z
    .boolean()
    .optional()
    .default(false)
    .describe("Accept subnet routes from other devices"),
  acceptDNS: z
    .boolean()
    .optional()
    .default(false)
    .describe("Accept DNS configuration from the network"),
  hostname: z
    .string()
    .optional()
    .describe("Set a custom hostname for this device"),
  advertiseRoutes: z
    .array(z.string())
    .optional()
    .describe("CIDR routes to advertise to other devices"),
  authKey: z
    .string()
    .optional()
    .describe("Authentication key for unattended setup"),
  loginServer: z.string().optional().describe("Custom coordination server URL"),
});

const PingPeerSchema = z.object({
  target: z.string().describe("Hostname or IP address of the target device"),
  count: z
    .number()
    .int()
    .min(1)
    .max(100)
    .optional()
    .default(4)
    .describe("Number of ping packets to send"),
});

// Tool handlers
async function getNetworkStatus(
  args: z.infer<typeof NetworkStatusSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Getting network status with format:", args.format);

    const result = await context.cli.getStatus();

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to get network status: ${result.error}`,
          },
        ],
        isError: true,
      };
    }

    const status = result.data!;

    if (args.format === "summary") {
      let output = `**Tailscale Network Status**\n\n`;
      output += `Version: ${status.Version}\n`;
      output += `Backend state: ${status.BackendState}\n`;
      output += `TUN interface: ${status.TUN ? "Active" : "Inactive"}\n`;
      output += `Tailscale IPs: ${(status.TailscaleIPs ?? []).join(", ")}\n\n`;

      output += `**This device:**\n`;
      output += `  - Hostname: ${status.Self.HostName}\n`;
      output += `  - DNS name: ${status.Self.DNSName}\n`;
      output += `  - OS: ${status.Self.OS}\n`;
      output += `  - IPs: ${status.Self.TailscaleIPs.join(", ")}\n`;
      output += `  - Online: ${status.Self.Online ? "🟢" : "🔴"}\n`;
      if (status.Self.ExitNode) {
        output += `  - Exit node: Yes\n`;
      }
      output += "\n";

      if (status.Peer && Object.keys(status.Peer).length > 0) {
        const peers = Object.values(status.Peer);
        output += `**Connected peers (${peers.length}):**\n`;
        for (const peer of peers) {
          const onlineStatus = peer.Online ? "🟢" : "🔴";
          output += `  ${onlineStatus} ${peer.HostName} (${peer.DNSName})\n`;
          output += `    - OS: ${peer.OS}\n`;
          output += `    - IPs: ${peer.TailscaleIPs.join(", ")}\n`;
          if (peer.LastSeen && peer.LastSeen !== "0001-01-01T00:00:00Z") {
            output += `    - Last seen: ${peer.LastSeen}\n`;
          }
          if (peer.ExitNode) {
            output += `    - Exit node: Yes\n`;
          }
          if (peer.Active) {
            output += `    - Active connection\n`;
          }
        }
      }

      return {
        content: [
          {
            type: "text",
            text: output,
          },
        ],
      };
    } else {
      // JSON format
      return {
        content: [
          {
            type: "text",
            text: JSON.stringify(status, null, 2),
          },
        ],
      };
    }
  } catch (error: any) {
    logger.error("Error getting network status:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error getting network status: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

async function connectNetwork(
  args: z.infer<typeof ConnectNetworkSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    const options = {
      acceptRoutes: args.acceptRoutes || false,
      acceptDNS: args.acceptDNS ?? false,
      hostname: args.hostname,
      advertiseRoutes: args.advertiseRoutes || [],
      authKey: args.authKey,
      loginServer: args.loginServer,
    };

    logger.debug("Connecting to Tailscale network with options:", options);

    const result = await context.cli.up(options);

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to connect to Tailscale: ${result.error}\n${
              result.stderr || ""
            }`,
          },
        ],
        isError: true,
      };
    }

    return {
      content: [
        {
          type: "text",
          text: `Successfully connected to Tailscale network\n\n${result.data}`,
        },
      ],
    };
  } catch (error: any) {
    logger.error("Error connecting to network:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error connecting to network: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

async function disconnectNetwork(
  args: any,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Disconnecting from Tailscale network");

    const result = await context.cli.down();

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to disconnect from Tailscale: ${result.error}\n${
              result.stderr || ""
            }`,
          },
        ],
        isError: true,
      };
    }

    return {
      content: [
        {
          type: "text",
          text: `Successfully disconnected from Tailscale network\n\n${result.data}`,
        },
      ],
    };
  } catch (error: any) {
    logger.error("Error disconnecting from network:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error disconnecting from network: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

async function pingPeer(
  args: z.infer<typeof PingPeerSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug(`Pinging ${args.target} (${args.count} packets)`);

    const result = await context.cli.ping(args.target, args.count);

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to ping ${args.target}: ${result.error}\n${
              result.stderr || ""
            }`,
          },
        ],
        isError: true,
      };
    }

    return {
      content: [
        {
          type: "text",
          text: `Ping results for ${args.target}:\n\n${result.data}`,
        },
      ],
    };
  } catch (error: any) {
    logger.error("Error pinging peer:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error pinging peer: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

async function getVersion(
  args: any,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Getting Tailscale version");

    const result = await context.cli.version();

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to get version: ${result.error}`,
          },
        ],
        isError: true,
      };
    }

    return {
      content: [
        {
          type: "text",
          text: `Tailscale version information:\n\n${result.data}`,
        },
      ],
    };
  } catch (error: any) {
    logger.error("Error getting version:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error getting version: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

// Export the tool module
export const networkTools: ToolModule = {
  tools: [
    {
      name: "get_network_status",
      description: "Get current network status from Tailscale CLI",
      inputSchema: NetworkStatusSchema,
      handler: getNetworkStatus,
    },
    {
      name: "connect_network",
      description: "Connect to the Tailscale network",
      inputSchema: ConnectNetworkSchema,
      handler: connectNetwork,
    },
    {
      name: "disconnect_network",
      description: "Disconnect from the Tailscale network",
      inputSchema: z.object({}),
      handler: disconnectNetwork,
    },
    {
      name: "ping_peer",
      description: "Ping a peer device",
      inputSchema: PingPeerSchema,
      handler: pingPeer,
    },
    {
      name: "get_version",
      description: "Get Tailscale version information",
      inputSchema: z.object({}),
      handler: getVersion,
    },
  ],
};
