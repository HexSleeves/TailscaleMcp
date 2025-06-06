#!/usr/bin/env node

/**
 * Tailscale MCP Server Test Script
 *
 * This script allows you to interact with the Tailscale MCP server locally
 * without requiring an AI assistant. It provides a simple CLI interface
 * to send tool requests and view responses.
 *
 * Usage:
 *   node test-mcp-server.js
 *
 * Environment Variables:
 *   TAILSCALE_API_KEY - Your Tailscale API key (optional for CLI-only operations)
 *   TAILSCALE_TAILNET - Your tailnet name (optional, defaults to current user's tailnet)
 *   MCP_START_TIMEOUT - Server startup timeout in milliseconds (default: 10000)
 *   MCP_LOG_FILE - Path to log file for session logging (optional, supports {timestamp} placeholder)
 *
 * Features:
 *   - Interactive CLI menu
 *   - Pre-defined common tool calls
 *   - Custom tool call input
 *   - Formatted response display
 *   - Environment variable validation
 *   - Optional file logging with timestamps
 */

// Load environment variables from .env file
import "dotenv/config";

import { spawn } from "node:child_process";
import { createInterface } from "node:readline";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import { writeFileSync, appendFileSync } from "node:fs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

class TailscaleMCPTester {
  constructor() {
    this.serverProcess = null;

    this.rl = createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    this.isConnected = false;

    // Initialize file logging if environment variable is set
    this.logFilePath = null;
    if (process.env.MCP_LOG_FILE) {
      const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
      this.logFilePath = process.env.MCP_LOG_FILE.replace(
        "{timestamp}",
        timestamp
      );

      // Create initial log file with header
      const header = `=== Tailscale MCP Server Test Log ===\nStarted: ${new Date().toISOString()}\n\n`;
      try {
        writeFileSync(this.logFilePath, header, "utf8");
        console.log(`📝 Logging to file: ${this.logFilePath}`);
      } catch (error) {
        console.error(`❌ Failed to create log file: ${error.message}`);
        this.logFilePath = null;
      }
    }
  }

  // ANSI color codes for better output formatting
  colors = {
    reset: "\x1b[0m",
    bright: "\x1b[1m",
    red: "\x1b[31m",
    green: "\x1b[32m",
    yellow: "\x1b[33m",
    blue: "\x1b[34m",
    magenta: "\x1b[35m",
    cyan: "\x1b[36m",
    white: "\x1b[37m",
  };

  log(message, color = "white") {
    const clr = this.colors[color] ?? this.colors.white;
    const coloredMessage = `${clr}${message}${this.colors.reset}`;
    console.log(coloredMessage);

    // Write to log file if enabled (without color codes)
    if (this.logFilePath) {
      const timestamp = new Date().toISOString();
      const logEntry = `[${timestamp}] ${message}\n`;
      try {
        appendFileSync(this.logFilePath, logEntry, "utf8");
      } catch (error) {
        console.error(`❌ Failed to write to log file: ${error.message}`);
      }
    }
  }

  async startServer() {
    return new Promise((resolve, reject) => {
      this.log("🚀 Starting Tailscale MCP Server...", "cyan");

      // Start the server process
      const serverPath = join(__dirname, "dist", "index.js");
      this.serverProcess = spawn("node", [serverPath], {
        stdio: ["pipe", "pipe", "pipe"],
        env: { ...process.env },
      });

      // Read timeout from environment variable, fallback to 10000 ms
      let timeoutMs = 10000;
      if (process.env.MCP_START_TIMEOUT) {
        const parsed = Number.parseInt(process.env.MCP_START_TIMEOUT, 10);
        if (Number.isFinite(parsed) && parsed > 0) {
          timeoutMs = parsed;
        }
      }

      const initTimeout = setTimeout(() => {
        reject(new Error("Server initialization timeout"));
      }, timeoutMs);

      // Handle server output
      this.serverProcess.stdout.on("data", (data) => {
        const output = data.toString();
        if (output.includes("Tailscale MCP Server started successfully")) {
          clearTimeout(initTimeout);
          this.isConnected = true;
          this.log("✅ Server started successfully!", "green");
          resolve();
        }
      });

      this.serverProcess.stderr.on("data", (data) => {
        const error = data.toString();
        this.log(`Server Error: ${error}`, "red");
      });

      this.serverProcess.on("error", (error) => {
        clearTimeout(initTimeout);
        reject(error);
      });

      this.serverProcess.on("exit", (code) => {
        this.isConnected = false;
        if (code !== 0) {
          this.log(`Server exited with code ${code}`, "red");
        }
      });
    });
  }

  async sendMCPRequest(method, params = {}) {
    if (!this.isConnected || !this.serverProcess) {
      throw new Error("Server not connected");
    }

    return new Promise((resolve, reject) => {
      const request = {
        jsonrpc: "2.0",
        id: Date.now(),
        method,
        params,
      };

      const requestStr = `${JSON.stringify(request)}\n`;

      let responseData = "";
      let responseTimeout = setTimeout(() => {
        reject(new Error("Request timeout"));
      }, 30000);

      const onData = (data) => {
        responseData += data.toString();

        // Try to parse complete JSON responses
        const lines = responseData.split("\n");
        for (const line of lines) {
          if (line.trim()) {
            try {
              const response = JSON.parse(line);
              if (response.id === request.id) {
                clearTimeout(responseTimeout);
                this.serverProcess.stdout.removeListener("data", onData);
                resolve(response);
                return;
              }
            } catch (e) {
              // Continue collecting data
            }
          }
        }
      };

      this.serverProcess.stdout.on("data", onData);
      this.serverProcess.stdin.write(requestStr);
    });
  }

  async listTools() {
    try {
      const response = await this.sendMCPRequest("tools/list");
      return response.result?.tools || [];
    } catch (error) {
      this.log(`Error listing tools: ${error.message}`, "red");
      return [];
    }
  }

  async callTool(name, args = {}) {
    try {
      const response = await this.sendMCPRequest("tools/call", {
        name,
        arguments: args,
      });

      if (response.error) {
        throw new Error(response.error.message || "Tool call failed");
      }

      return response.result;
    } catch (error) {
      this.log(`Error calling tool ${name}: ${error.message}`, "red");
      return null;
    }
  }

  formatToolResponse(result) {
    if (!result) return "No result";

    if (result.isError) {
      this.log("❌ Tool Error:", "red");
    } else {
      this.log("✅ Tool Result:", "green");
    }

    if (result.content && Array.isArray(result.content)) {
      result.content.forEach((item) => {
        if (item.type === "text") {
          console.log(item.text);
        }
      });
    } else {
      console.log(JSON.stringify(result, null, 2));
    }
  }

  async showMainMenu() {
    console.clear();
    this.log("🔧 Tailscale MCP Server Test Interface", "cyan");
    this.log("=====================================", "cyan");
    console.log();

    // Check environment
    this.checkEnvironment();
    console.log();

    this.log("Available Options:", "yellow");
    console.log("1. List all available tools");
    console.log("2. Get Tailscale version");
    console.log("3. Get network status");
    console.log("4. List devices");
    console.log("5. Ping a peer");
    console.log("6. Get tailnet info");
    console.log("7. Custom tool call");
    console.log("8. Show example tool calls");
    console.log("9. Exit (q)");
    console.log();

    const choice = await this.prompt("Select an option (1-9, q to quit): ");
    return choice;
  }

  checkEnvironment() {
    this.log("Environment Check:", "yellow");

    const apiKey = process.env.TAILSCALE_API_KEY;
    const tailnet = process.env.TAILSCALE_TAILNET;

    if (apiKey) {
      this.log(
        `✅ TAILSCALE_API_KEY: Set (${apiKey.substring(0, 8)}...)`,
        "green"
      );
    } else {
      this.log(
        "⚠️  TAILSCALE_API_KEY: Not set (API operations will fail)",
        "yellow"
      );
    }

    if (tailnet) {
      this.log(`✅ TAILSCALE_TAILNET: ${tailnet}`, "green");
    } else {
      this.log("ℹ️  TAILSCALE_TAILNET: Using default (-)", "blue");
    }

    // Show logging status
    if (this.logFilePath) {
      this.log(`📝 LOGGING: Enabled (${this.logFilePath})`, "green");
    } else if (process.env.MCP_LOG_FILE) {
      this.log("⚠️  LOGGING: Failed to initialize log file", "yellow");
    } else {
      this.log("ℹ️  LOGGING: Disabled (set MCP_LOG_FILE to enable)", "blue");
    }
  }

  async handleMenuChoice(choice) {
    switch (choice.trim()) {
      case "1":
        await this.listAllTools();
        break;
      case "2":
        await this.getVersion();
        break;
      case "3":
        await this.getNetworkStatus();
        break;
      case "4":
        await this.listDevices();
        break;
      case "5":
        await this.pingPeer();
        break;
      case "6":
        await this.getTailnetInfo();
        break;
      case "7":
        await this.customToolCall();
        break;
      case "8":
        await this.showExamples();
        break;
      case "9":
      case "q":
        await this.exit();
        return false;
      default:
        this.log("Invalid option. Please try again.", "red");
        await this.pause();
        break;
    }
    return true;
  }

  async listAllTools() {
    this.log("\n📋 Listing all available tools...", "cyan");
    const tools = await this.listTools();

    if (tools.length === 0) {
      this.log("No tools available", "yellow");
      return;
    }

    console.log();
    tools.forEach((tool, index) => {
      this.log(`${index + 1}. ${tool.name}`, "bright");
      console.log(`   Description: ${tool.description}`);
      if (tool.inputSchema && tool.inputSchema.properties) {
        const props = Object.keys(tool.inputSchema.properties);
        if (props.length > 0) {
          console.log(`   Parameters: ${props.join(", ")}`);
        }
      }
      console.log();
    });

    await this.pause();
  }

  async getVersion() {
    this.log("\n🔍 Getting Tailscale version...", "cyan");
    const result = await this.callTool("get_version");
    this.formatToolResponse(result);
    await this.pause();
  }

  async getNetworkStatus() {
    this.log("\n🌐 Getting network status...", "cyan");
    const format =
      (await this.prompt("Format (json/summary) [summary]: ")) || "summary";
    const result = await this.callTool("get_network_status", { format });
    this.formatToolResponse(result);
    await this.pause();
  }

  async listDevices() {
    this.log("\n📱 Listing devices...", "cyan");
    const result = await this.callTool("list_devices");
    this.formatToolResponse(result);
    await this.pause();
  }

  async pingPeer() {
    this.log("\n🏓 Ping a peer...", "cyan");
    const target = await this.prompt("Enter target hostname or IP: ");
    if (!target.trim()) {
      this.log("Target is required", "red");
      await this.pause();
      return;
    }

    const countStr = (await this.prompt("Number of pings [4]: ")) || "4";
    const count = parseInt(countStr) || 4;

    const result = await this.callTool("ping_peer", {
      target: target.trim(),
      count,
    });
    this.formatToolResponse(result);
    await this.pause();
  }

  async getTailnetInfo() {
    this.log("\n🏢 Getting tailnet information...", "cyan");
    const includeDetails = await this.prompt(
      "Include detailed info? (y/n) [n]: "
    );
    const result = await this.callTool("get_tailnet_info", {
      includeDetails: includeDetails.toLowerCase().startsWith("y"),
    });
    this.formatToolResponse(result);
    await this.pause();
  }

  async customToolCall() {
    this.log("\n🛠️  Custom tool call...", "cyan");

    const tools = await this.listTools();
    if (tools.length === 0) {
      this.log("No tools available", "yellow");
      await this.pause();
      return;
    }

    console.log("\nAvailable tools:");
    tools.forEach((tool, index) => {
      console.log(`${index + 1}. ${tool.name}`);
    });

    const toolName = await this.prompt("\nEnter tool name: ");
    if (!toolName.trim()) {
      this.log("Tool name is required", "red");
      await this.pause();
      return;
    }

    const argsStr = await this.prompt(
      "Enter arguments as JSON (or press Enter for empty): "
    );
    let args = {};

    if (argsStr.trim()) {
      try {
        args = JSON.parse(argsStr);
      } catch (error) {
        this.log(`Invalid JSON: ${error.message}`, "red");
        await this.pause();
        return;
      }
    }

    const result = await this.callTool(toolName.trim(), args);
    this.formatToolResponse(result);
    await this.pause();
  }

  async showExamples() {
    this.log("\n📚 Example Tool Calls", "cyan");
    console.log("====================");
    console.log();

    const examples = [
      {
        name: "get_version",
        args: {},
        description: "Get Tailscale version information",
      },
      {
        name: "get_network_status",
        args: { format: "summary" },
        description: "Get network status in summary format",
      },
      {
        name: "list_devices",
        args: {},
        description: "List all devices in the tailnet",
      },
      {
        name: "ping_peer",
        args: { target: "hostname-or-ip", count: 4 },
        description: "Ping a peer device",
      },
      {
        name: "connect_network",
        args: { acceptRoutes: true, acceptDNS: true },
        description: "Connect to Tailscale with route and DNS acceptance",
      },
      {
        name: "device_action",
        args: { deviceId: "device-id", action: "authorize" },
        description: "Authorize a device",
      },
      {
        name: "manage_routes",
        args: {
          deviceId: "device-id",
          routes: ["192.168.1.0/24"],
          action: "enable",
        },
        description: "Enable subnet routes for a device",
      },
    ];

    examples.forEach((example, index) => {
      this.log(`${index + 1}. ${example.name}`, "bright");
      console.log(`   Description: ${example.description}`);
      console.log(`   Arguments: ${JSON.stringify(example.args, null, 2)}`);
      console.log();
    });

    await this.pause();
  }

  async prompt(question) {
    return new Promise((resolve) => {
      this.rl.question(question, resolve);
    });
  }

  async pause() {
    await this.prompt("\nPress Enter to continue...");
  }

  async exit() {
    this.log("\n👋 Shutting down...", "cyan");

    if (this.serverProcess) {
      this.serverProcess.kill("SIGTERM");

      // Wait a bit for graceful shutdown
      await new Promise((resolve) => setTimeout(resolve, 1000));

      if (!this.serverProcess.killed) {
        this.serverProcess.kill("SIGKILL");
      }
    }

    this.rl.close();
    this.log("Goodbye!", "green");
    process.exit(0);
  }

  async runMainLoop() {
    while (true) {
      const choice = await this.showMainMenu();
      const shouldContinue = await this.handleMenuChoice(choice);
      if (shouldContinue === false) {
        break;
      }
    }
  }

  async run() {
    try {
      // Check if the server build exists
      const serverPath = join(__dirname, "dist", "index.js");
      try {
        await import("fs").then((fs) => fs.promises.access(serverPath));
      } catch (error) {
        this.log(
          '❌ Server build not found. Please run "npm run build" first.',
          "red"
        );
        process.exit(1);
      }

      await this.startServer();
      await this.runMainLoop();
    } catch (error) {
      this.log(`❌ Failed to start: ${error.message}`, "red");
      process.exit(1);
    }
  }
}

// Handle process termination
process.on("SIGINT", async () => {
  console.log("\n\n🛑 Received SIGINT, shutting down gracefully...");
  process.exit(0);
});

process.on("SIGTERM", async () => {
  console.log("\n\n🛑 Received SIGTERM, shutting down gracefully...");
  process.exit(0);
});

// Start the tester
const tester = new TailscaleMCPTester();
tester.run().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
