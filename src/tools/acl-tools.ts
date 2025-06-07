import type { CallToolResult } from "@modelcontextprotocol/sdk/types.js";
import { z } from "zod/v4";
import { logger } from "../logger.js";
import type { ToolContext, ToolModule } from "./index.js";

// Schemas
const ACLSchema = z.object({
  operation: z
    .enum(["get", "update", "validate"])
    .describe("ACL operation to perform"),
  aclConfig: z
    .object({
      acls: z
        .array(
          z.object({
            action: z.enum(["accept", "drop"]),
            src: z.array(z.string()),
            dst: z.array(z.string()),
          })
        )
        .optional()
        .describe("Access control rules"),
      groups: z
        .record(z.string(), z.array(z.string()))
        .optional()
        .describe("User groups definition"),
      tagOwners: z
        .record(z.string(), z.array(z.string()))
        .optional()
        .describe("Tag ownership mapping"),
    })
    .optional()
    .describe("ACL configuration (required for update/validate operations)"),
});

const DNSSchema = z.object({
  operation: z
    .enum([
      "get_nameservers",
      "set_nameservers",
      "get_preferences",
      "set_preferences",
      "get_searchpaths",
      "set_searchpaths",
    ])
    .describe("DNS operation to perform"),
  nameservers: z
    .array(z.string())
    .optional()
    .describe("DNS nameservers (for set_nameservers operation)"),
  magicDNS: z
    .boolean()
    .optional()
    .describe("Enable/disable MagicDNS (for set_preferences operation)"),
  searchPaths: z
    .array(z.string())
    .optional()
    .describe("DNS search paths (for set_searchpaths operation)"),
});

const KeyManagementSchema = z.object({
  operation: z
    .enum(["list", "create", "delete"])
    .describe("Key management operation"),
  keyConfig: z
    .object({
      description: z.string().optional(),
      expirySeconds: z.number().optional(),
      capabilities: z
        .object({
          devices: z
            .object({
              create: z
                .object({
                  reusable: z.boolean().optional(),
                  ephemeral: z.boolean().optional(),
                  preauthorized: z.boolean().optional(),
                  tags: z.array(z.string()).optional(),
                })
                .optional(),
            })
            .optional(),
        })
        .optional(),
    })
    .optional()
    .describe("Key configuration (for create operation)"),
  keyId: z
    .string()
    .optional()
    .describe("Authentication key ID (for delete operation)"),
});

const NetworkLockSchema = z.object({
  operation: z
    .enum(["status", "enable", "disable", "add_key", "remove_key", "list_keys"])
    .describe("Network lock operation to perform"),
  publicKey: z
    .string()
    .optional()
    .describe("Public key for add/remove operations"),
  keyId: z.string().optional().describe("Key ID for remove operations"),
});

const PolicyFileSchema = z.object({
  operation: z
    .enum(["get", "update", "test_access"])
    .describe("Policy file operation to perform"),
  policy: z
    .string()
    .optional()
    .describe("Policy content (HuJSON format) for update operation"),
  testRequest: z
    .object({
      src: z.string(),
      dst: z.string(),
      proto: z.string().optional(),
    })
    .optional()
    .describe("Access test parameters for test_access operation"),
});

// Tool handlers
async function manageACL(
  args: z.infer<typeof ACLSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Managing ACL configuration:", args);

    switch (args.operation) {
      case "get": {
        const result = await context.api.getACL();
        if (!result.success) {
          return {
            content: [
              { type: "text", text: `Failed to get ACL: ${result.error}` },
            ],
            isError: true,
          };
        }
        return {
          content: [
            {
              type: "text",
              text: `Current ACL configuration:\n\n${
                typeof result.data === "string"
                  ? result.data
                  : JSON.stringify(result.data, null, 2)
              }`,
            },
          ],
        };
      }

      case "update": {
        if (!args.aclConfig) {
          return {
            content: [
              {
                type: "text",
                text: "ACL configuration is required for update operation",
              },
            ],
            isError: true,
          };
        }

        const aclString = JSON.stringify(args.aclConfig, null, 2);
        const result = await context.api.updateACL(aclString);

        if (!result.success) {
          return {
            content: [
              { type: "text", text: `Failed to update ACL: ${result.error}` },
            ],
            isError: true,
          };
        }

        return {
          content: [
            { type: "text", text: "ACL configuration updated successfully" },
          ],
        };
      }

      case "validate": {
        if (!args.aclConfig) {
          return {
            content: [
              {
                type: "text",
                text: "ACL configuration is required for validation",
              },
            ],
            isError: true,
          };
        }

        const aclString = JSON.stringify(args.aclConfig, null, 2);
        const result = await context.api.validateACL(aclString);

        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `ACL validation failed: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [{ type: "text", text: "ACL configuration is valid" }],
        };
      }

      default:
        return {
          content: [
            {
              type: "text",
              text: "Invalid ACL operation. Use: get, update, or validate",
            },
          ],
          isError: true,
        };
    }
  } catch (error: any) {
    logger.error("Error managing ACL:", error);
    return {
      content: [{ type: "text", text: `Error managing ACL: ${error.message}` }],
      isError: true,
    };
  }
}

async function manageDNS(
  args: z.infer<typeof DNSSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Managing DNS configuration:", args);

    switch (args.operation) {
      case "get_nameservers": {
        const result = await context.api.getDNSNameservers();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to get DNS nameservers: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        const nameservers = result.data?.dns || [];
        return {
          content: [
            {
              type: "text",
              text: `DNS Nameservers:\n${
                nameservers.length > 0
                  ? nameservers.map((ns) => `  - ${ns}`).join("\n")
                  : "  No custom nameservers configured"
              }`,
            },
          ],
        };
      }

      case "set_nameservers": {
        if (!args.nameservers) {
          return {
            content: [
              {
                type: "text",
                text: "Nameservers array is required for set_nameservers operation",
              },
            ],
            isError: true,
          };
        }

        const result = await context.api.setDNSNameservers(args.nameservers);
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to set DNS nameservers: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `DNS nameservers updated to: ${args.nameservers.join(
                ", "
              )}`,
            },
          ],
        };
      }

      case "get_preferences": {
        const result = await context.api.getDNSPreferences();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to get DNS preferences: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `DNS Preferences:\n  MagicDNS: ${
                result.data?.magicDNS ? "Enabled" : "Disabled"
              }`,
            },
          ],
        };
      }

      case "set_preferences": {
        if (args.magicDNS === undefined) {
          return {
            content: [
              {
                type: "text",
                text: "magicDNS boolean is required for set_preferences operation",
              },
            ],
            isError: true,
          };
        }

        const result = await context.api.setDNSPreferences(args.magicDNS);
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to set DNS preferences: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `MagicDNS ${args.magicDNS ? "enabled" : "disabled"}`,
            },
          ],
        };
      }

      case "get_searchpaths": {
        const result = await context.api.getDNSSearchPaths();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to get DNS search paths: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        const searchPaths = result.data?.searchPaths || [];
        return {
          content: [
            {
              type: "text",
              text: `DNS Search Paths:\n${
                searchPaths.length > 0
                  ? searchPaths.map((path) => `  - ${path}`).join("\n")
                  : "  No search paths configured"
              }`,
            },
          ],
        };
      }

      case "set_searchpaths": {
        if (!args.searchPaths) {
          return {
            content: [
              {
                type: "text",
                text: "searchPaths array is required for set_searchpaths operation",
              },
            ],
            isError: true,
          };
        }

        const result = await context.api.setDNSSearchPaths(args.searchPaths);
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to set DNS search paths: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `DNS search paths updated to: ${args.searchPaths.join(
                ", "
              )}`,
            },
          ],
        };
      }

      default:
        return {
          content: [
            {
              type: "text",
              text: "Invalid DNS operation. Use: get_nameservers, set_nameservers, get_preferences, set_preferences, get_searchpaths, set_searchpaths",
            },
          ],
          isError: true,
        };
    }
  } catch (error: any) {
    logger.error("Error managing DNS:", error);
    return {
      content: [{ type: "text", text: `Error managing DNS: ${error.message}` }],
      isError: true,
    };
  }
}

async function manageKeys(
  args: z.infer<typeof KeyManagementSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Managing authentication keys:", args);

    switch (args.operation) {
      case "list": {
        const result = await context.api.listAuthKeys();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to list auth keys: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        const keys = result.data?.keys || [];
        if (keys.length === 0) {
          return {
            content: [{ type: "text", text: "No authentication keys found" }],
          };
        }

        const keyList = keys
          .map((key: any, index: number) => {
            return `**Key ${index + 1}**
  - ID: ${key.id}
  - Description: ${key.description || "No description"}
  - Created: ${key.created}
  - Expires: ${key.expires}
  - Revoked: ${key.revoked ? "Yes" : "No"}
  - Reusable: ${key.capabilities?.devices?.create?.reusable ? "Yes" : "No"}
  - Preauthorized: ${
    key.capabilities?.devices?.create?.preauthorized ? "Yes" : "No"
  }`;
          })
          .join("\n\n");

        return {
          content: [
            {
              type: "text",
              text: `Found ${keys.length} authentication keys:\n\n${keyList}`,
            },
          ],
        };
      }

      case "create": {
        if (!args.keyConfig) {
          return {
            content: [
              {
                type: "text",
                text: "Key configuration is required for create operation",
              },
            ],
            isError: true,
          };
        }

        const result = await context.api.createAuthKey(args.keyConfig);
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to create auth key: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `Authentication key created successfully:
  - ID: ${result.data?.id}
  - Key: ${result.data?.key}
  - Description: ${result.data?.description || "No description"}`,
            },
          ],
        };
      }

      case "delete": {
        if (!args.keyId) {
          return {
            content: [
              {
                type: "text",
                text: "Key ID is required for delete operation",
              },
            ],
            isError: true,
          };
        }

        const result = await context.api.deleteAuthKey(args.keyId);
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to delete auth key: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `Authentication key ${args.keyId} deleted successfully`,
            },
          ],
        };
      }

      default:
        return {
          content: [
            {
              type: "text",
              text: "Invalid key operation. Use: list, create, or delete",
            },
          ],
          isError: true,
        };
    }
  } catch (error: any) {
    logger.error("Error managing keys:", error);
    return {
      content: [
        { type: "text", text: `Error managing keys: ${error.message}` },
      ],
      isError: true,
    };
  }
}

async function manageNetworkLock(
  args: z.infer<typeof NetworkLockSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Managing network lock:", args);

    switch (args.operation) {
      case "status": {
        const result = await context.api.getNetworkLockStatus();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to get network lock status: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        const status = result.data;
        return {
          content: [
            {
              type: "text",
              text: `Network Lock Status:
  - Enabled: ${status?.enabled ? "Yes" : "No"}
  - Node Key: ${status?.nodeKey || "Not available"}
  - Trusted Keys: ${status?.trustedKeys?.length || 0}`,
            },
          ],
        };
      }

      case "enable": {
        const result = await context.api.enableNetworkLock();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to enable network lock: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `Network lock enabled successfully. Key: ${
                result.data?.key || "Generated"
              }`,
            },
          ],
        };
      }

      case "disable": {
        const result = await context.api.disableNetworkLock();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to disable network lock: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            { type: "text", text: "Network lock disabled successfully" },
          ],
        };
      }

      default:
        return {
          content: [
            {
              type: "text",
              text: "Invalid network lock operation. Use: status, enable, or disable",
            },
          ],
          isError: true,
        };
    }
  } catch (error: any) {
    logger.error("Error managing network lock:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error managing network lock: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

async function managePolicyFile(
  args: z.infer<typeof PolicyFileSchema>,
  context: ToolContext
): Promise<CallToolResult> {
  try {
    logger.debug("Managing policy file:", args);

    switch (args.operation) {
      case "get": {
        const result = await context.api.getPolicyFile();
        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to get policy file: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        return {
          content: [
            {
              type: "text",
              text: `Policy File (HuJSON format):\n\n${result.data}`,
            },
          ],
        };
      }

      case "test_access": {
        if (!args.testRequest) {
          return {
            content: [
              {
                type: "text",
                text: "Test request parameters are required for test_access operation",
              },
            ],
            isError: true,
          };
        }

        const { src, dst, proto } = args.testRequest;
        const result = await context.api.testACLAccess(src, dst, proto);

        if (!result.success) {
          return {
            content: [
              {
                type: "text",
                text: `Failed to test ACL access: ${result.error}`,
              },
            ],
            isError: true,
          };
        }

        const testResult = result.data;
        return {
          content: [
            {
              type: "text",
              text: `ACL Access Test Result:
  - Source: ${src}
  - Destination: ${dst}
  - Protocol: ${proto || "any"}
  - Result: ${testResult?.allowed ? "ALLOWED" : "DENIED"}
  - Rule: ${testResult?.rule || "No matching rule"}
  - Match: ${testResult?.match || "N/A"}`,
            },
          ],
        };
      }

      default:
        return {
          content: [
            {
              type: "text",
              text: "Invalid policy operation. Use: get or test_access",
            },
          ],
          isError: true,
        };
    }
  } catch (error: any) {
    logger.error("Error managing policy file:", error);
    return {
      content: [
        {
          type: "text",
          text: `Error managing policy file: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
}

// Export the tool module
export const aclTools: ToolModule = {
  tools: [
    {
      name: "manage_acl",
      description: "Manage Tailscale Access Control Lists (ACLs)",
      inputSchema: ACLSchema,
      handler: manageACL,
    },
    {
      name: "manage_dns",
      description: "Manage Tailscale DNS configuration",
      inputSchema: DNSSchema,
      handler: manageDNS,
    },
    {
      name: "manage_keys",
      description: "Manage Tailscale authentication keys",
      inputSchema: KeyManagementSchema,
      handler: manageKeys,
    },
    {
      name: "manage_network_lock",
      description:
        "Manage Tailscale network lock (key authority) for enhanced security",
      inputSchema: NetworkLockSchema,
      handler: manageNetworkLock,
    },
    {
      name: "manage_policy_file",
      description: "Manage policy files and test ACL access rules",
      inputSchema: PolicyFileSchema,
      handler: managePolicyFile,
    },
  ],
};
