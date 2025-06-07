import type { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod/v4";
import { logger } from "../logger.js";
import type { ToolContext, ToolModule } from "./index.js";

// Schemas
const ListDevicesSchema = z.object({
  includeRoutes: z
    .boolean()
    .optional()
    .default(false)
    .describe("Include route information for each device"),
});

const DeviceActionSchema = z.object({
  deviceId: z.string().describe("The ID of the device to act on"),
  action: z
    .enum(["authorize", "deauthorize", "delete", "expire-key"])
    .describe("The action to perform on the device"),
});

const ManageRoutesSchema = z.object({
  deviceId: z.string().describe("The ID of the device"),
  routes: z.array(z.string()).describe("Array of CIDR routes to manage"),
  action: z
    .enum(["enable", "disable"])
    .describe("Whether to enable or disable the routes"),
});

// Tool handlers
async function listDevices(
  args: z.infer<typeof ListDevicesSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Listing devices with options:", args);

    const result = await context.api.listDevices();

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to list devices: ${result.error}`,
          },
        ],
        isError: true,
      };
    }

    const devices = result.data!;
    let output = `Found ${devices.length} devices:\n\n`;

    for (const device of devices) {
      output += `**${device.name}** (${device.hostname})\n`;
      output += `  - ID: ${device.id}\n`;
      output += `  - OS: ${device.os}\n`;
      output += `  - Addresses: ${device.addresses.join(", ")}\n`;
      output += `  - Authorized: ${device.authorized ? "✅" : "❌"}\n`;
      output += `  - Last seen: ${device.lastSeen}\n`;
      output += `  - Client version: ${device.clientVersion}\n`;

      if (
        args.includeRoutes &&
        Array.isArray(device.advertisedRoutes) &&
        device.advertisedRoutes.length > 0
      ) {
        output += `  - Advertised routes: ${device.advertisedRoutes.join(
          ", "
        )}\n`;
        output += `  - Enabled routes: ${
          Array.isArray(device.enabledRoutes)
            ? device.enabledRoutes.join(", ")
            : "—"
        }\n`;
      }

      output += "\n";
    }

    return {
      content: [
        {
          type: "text",
          text: output,
        },
      ],
    };
  } catch (error: any) {
    logger.error("Error listing devices:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error listing devices: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

async function deviceAction(
  args: z.infer<typeof DeviceActionSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Performing device action:", args);

    let result;
    switch (args.action) {
      case "authorize":
        result = await context.api.authorizeDevice(args.deviceId);
        break;
      case "deauthorize":
        result = await context.api.deauthorizeDevice(args.deviceId);
        break;
      case "delete":
        result = await context.api.deleteDevice(args.deviceId);
        break;
      case "expire-key":
        result = await context.api.expireDeviceKey(args.deviceId);
        break;
      default:
        return {
          content: [
            {
              type: "text",
              text: `Unknown action: ${args.action}`,
            },
          ],
          isError: true,
        };
    }

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to perform device action: ${result.error}`,
          },
        ],
        isError: true,
      };
    }

    return {
      content: [
        {
          type: "text",
          text: `Successfully performed action "${args.action}" on device ${args.deviceId}`,
        },
      ],
    };
  } catch (error: any) {
    logger.error("Error performing device action:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error performing device action: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

async function manageRoutes(
  args: z.infer<typeof ManageRoutesSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Managing routes:", args);

    let result;
    if (args.action === "enable") {
      result = await context.api.enableDeviceRoutes(args.deviceId, args.routes);
    } else {
      result = await context.api.disableDeviceRoutes(
        args.deviceId,
        args.routes
      );
    }

    if (!result.success) {
      return {
        content: [
          {
            type: "text",
            text: `Failed to manage routes: ${result.error}`,
          },
        ],
        isError: true,
      };
    }

    return {
      content: [
        {
          type: "text",
          text: `Successfully ${args.action}d routes ${args.routes.join(
            ", "
          )} for device ${args.deviceId}`,
        },
      ],
    };
  } catch (error: any) {
    logger.error("Error managing routes:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error managing routes: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

// Export the tool module
export const deviceTools: ToolModule = {
  tools: [
    {
      name: "list_devices",
      description: "List all devices in the Tailscale network",
      inputSchema: ListDevicesSchema,
      handler: listDevices,
    },
    {
      name: "device_action",
      description: "Perform actions on a specific device",
      inputSchema: DeviceActionSchema,
      handler: deviceAction,
    },
    {
      name: "manage_routes",
      description: "Enable or disable routes for a device",
      inputSchema: ManageRoutesSchema,
      handler: manageRoutes,
    },
  ],
};
