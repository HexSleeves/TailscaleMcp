import { execFile } from "child_process";
import { promisify } from "util";
import {
  TailscaleCLIStatus,
  TailscaleCLIStatusSchema,
  CLIResponse,
} from "../types.js";
import { logger } from "../logger.js";

const execFileAsync = promisify(execFile);

// Validate target format (hostname, IP, or Tailscale node name)
// Hostname/IP pattern: no leading/trailing dots or hyphens, no consecutive dots
export const VALID_TARGET_PATTERN =
  /^(([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*)|([0-9a-fA-F:]+))$/;

// CIDR validation
export const cidrPattern =
  /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$|^([0-9a-fA-F:]+)\/\d{1,3}$/;

export class TailscaleCLI {
  private cliPath: string;

  constructor(cliPath: string = "tailscale") {
    this.cliPath = cliPath;
  }

  private validateTarget(target: string): void {
    if (!target || typeof target !== "string") {
      throw new Error("Invalid target specified");
    }

    // Comprehensive validation to prevent command injection
    const dangerousChars = [
      ";",
      "&",
      "|",
      "`",
      "$",
      "(",
      ")",
      "{",
      "}",
      "[",
      "]",
      "<",
      ">",
      "\\",
      "'",
      '"',
    ];
    for (const char of dangerousChars) {
      if (target.includes(char)) {
        throw new Error(`Invalid character '${char}' in target`);
      }
    }

    // Additional validation for common patterns
    if (
      target.includes("..") ||
      target.startsWith("/") ||
      target.includes("~")
    ) {
      throw new Error("Invalid path patterns in target");
    }

    // Validate target format (hostname, IP, or Tailscale node name)
    // Hostname/IP pattern: no leading/trailing dots or hyphens, no consecutive dots
    if (!VALID_TARGET_PATTERN.test(target)) {
      throw new Error("Target contains invalid characters");
    }

    // Length validation
    if (target.length > 253) {
      // DNS hostname max length
      throw new Error("Target too long");
    }
  }

  private validateStringInput(input: string, fieldName: string): void {
    if (typeof input !== "string") {
      throw new Error(`${fieldName} must be a string`);
    }

    // Check for dangerous characters
    const dangerousChars = [
      ";",
      "&",
      "|",
      "`",
      "$",
      "(",
      ")",
      "{",
      "}",
      "<",
      ">",
      "\\",
    ];
    for (const char of dangerousChars) {
      if (input.includes(char)) {
        throw new Error(`Invalid character '${char}' in ${fieldName}`);
      }
    }

    // Length validation
    if (input.length > 1000) {
      throw new Error(`${fieldName} too long`);
    }
  }

  private validateRoutes(routes: string[]): void {
    if (!Array.isArray(routes)) {
      throw new Error("Routes must be an array");
    }

    for (const route of routes) {
      if (typeof route !== "string") {
        throw new Error("Each route must be a string");
      }

      // Basic CIDR validation
      // More precise CIDR validation
      if (
        !cidrPattern.test(route) &&
        route !== "0.0.0.0/0" &&
        route !== "::/0"
      ) {
        throw new Error(`Invalid route format: ${route}`);
      }
    }
  }

  /**
   * Execute a Tailscale CLI command
   */
  private async executeCommand(args: string[]): Promise<CLIResponse<string>> {
    try {
      // Validate all arguments
      for (const arg of args) {
        if (typeof arg !== "string") {
          throw new Error("All command arguments must be strings");
        }

        // Basic validation for each argument
        if (arg.length > 1000) {
          throw new Error("Command argument too long");
        }
      }

      logger.debug(`Executing: ${this.cliPath} ${args.join(" ")}`);

      const { stdout, stderr } = await execFileAsync(this.cliPath, args, {
        encoding: "utf8",
        maxBuffer: 1024 * 1024 * 10, // 10MB buffer limit
        timeout: 30000, // 30 second timeout
        windowsHide: true, // Hide window on Windows
        killSignal: "SIGTERM", // Graceful termination signal
      });

      if (stderr && stderr.trim()) {
        logger.warn("CLI stderr:", stderr);
      }

      return {
        success: true,
        data: stdout.trim(),
        stderr: stderr?.trim(),
      };
    } catch (error: any) {
      logger.error("CLI command failed:", error);

      return {
        success: false,
        error: error.message,
        stderr: error.stderr,
      };
    }
  }

  /**
   * Get Tailscale status
   */
  async getStatus(): Promise<CLIResponse<TailscaleCLIStatus>> {
    const result = await this.executeCommand(["status", "--json"]);

    if (!result.success) {
      return {
        success: false,
        error: result.error || "Unknown error",
        stderr: result.stderr,
      };
    }

    try {
      const statusData = JSON.parse(result.data!);
      const validatedStatus = TailscaleCLIStatusSchema.parse(statusData);

      return {
        success: true,
        data: validatedStatus,
      };
    } catch (error: any) {
      logger.error("Failed to parse status JSON:", error);
      return {
        success: false,
        error: `Failed to parse status data: ${error.message}`,
      };
    }
  }

  /**
   * Get list of peers
   */
  async listPeers(): Promise<CLIResponse<string[]>> {
    const statusResult = await this.getStatus();

    if (!statusResult.success) {
      return {
        success: false,
        error: statusResult.error || "Unknown error",
        stderr: statusResult.stderr,
      };
    }

    const peers = statusResult.data?.Peer
      ? Object.values(statusResult.data.Peer)
          .map((p) => (p as any).HostName)
          .filter(
            (hostname): hostname is string => typeof hostname === "string"
          )
      : [];

    return {
      success: true,
      data: peers,
    };
  }

  /**
   * Connect to Tailscale network
   */
  async up(
    options: {
      loginServer?: string;
      acceptRoutes?: boolean;
      acceptDns?: boolean;
      hostname?: string;
      advertiseRoutes?: string[];
      authKey?: string;
    } = {}
  ): Promise<CLIResponse<string>> {
    const args = ["up"];

    if (options.loginServer) {
      this.validateStringInput(options.loginServer, "loginServer");
      args.push("--login-server", options.loginServer);
    }

    if (options.acceptRoutes) {
      args.push("--accept-routes");
    }

    if (options.acceptDns) {
      args.push("--accept-dns");
    }

    if (options.hostname) {
      this.validateStringInput(options.hostname, "hostname");
      args.push("--hostname", options.hostname);
    }

    if (options.advertiseRoutes && options.advertiseRoutes.length > 0) {
      this.validateRoutes(options.advertiseRoutes);
      args.push("--advertise-routes", options.advertiseRoutes.join(","));
    }

    if (options.authKey) {
      this.validateStringInput(options.authKey, "authKey");
      // Pass auth key directly as argument since execFile handles it securely
      // The auth key won't be exposed in shell command history
      // Changed from info to debug
      logger.debug("Auth key passed securely via execFile");
      args.push("--authkey", options.authKey);
    }

    return await this.executeCommand(args);
  }

  /**
   * Disconnect from Tailscale network
   */
  async down(): Promise<CLIResponse<string>> {
    return await this.executeCommand(["down"]);
  }

  /**
   * Ping a peer
   */
  private static readonly MIN_PING_COUNT = 1;
  private static readonly MAX_PING_COUNT = 100;

  async ping(target: string, count: number = 4): Promise<CLIResponse<string>> {
    this.validateTarget(target);

    if (
      !Number.isInteger(count) ||
      count < TailscaleCLI.MIN_PING_COUNT ||
      count > TailscaleCLI.MAX_PING_COUNT
    ) {
      throw new Error(
        `Count must be an integer between ${TailscaleCLI.MIN_PING_COUNT} and ${TailscaleCLI.MAX_PING_COUNT}`
      );
    }

    return await this.executeCommand(["ping", target, "-c", count.toString()]);
  }

  /**
   * Get network check information
   */
  async netcheck(): Promise<CLIResponse<string>> {
    return await this.executeCommand(["netcheck"]);
  }

  /**
   * Get version information
   */
  async version(): Promise<CLIResponse<string>> {
    return await this.executeCommand(["version"]);
  }

  /**
   * Logout from Tailscale
   */
  async logout(): Promise<CLIResponse<string>> {
    return await this.executeCommand(["logout"]);
  }

  /**
   * Set exit node
   */
  async setExitNode(nodeId?: string): Promise<CLIResponse<string>> {
    const args = ["set"];

    if (nodeId) {
      this.validateTarget(nodeId);
      args.push("--exit-node", nodeId);
    } else {
      args.push("--exit-node="); // Clear exit node with empty value
    }

    return await this.executeCommand(args);
  }

  /**
   * Enable/disable shields up mode
   */
  async setShieldsUp(enabled: boolean): Promise<CLIResponse<string>> {
    return await this.executeCommand([
      "set",
      "--shields-up",
      enabled ? "true" : "false",
    ]);
  }

  /**
   * Check if CLI is available
   */
  async isAvailable(): Promise<boolean> {
    try {
      const result = await this.executeCommand(["version"]);
      return result.success;
    } catch (error) {
      logger.error("tailscale CLI not found:", error);
      return false;
    }
  }
}

// Export default instance
export const tailscaleCLI = new TailscaleCLI();
