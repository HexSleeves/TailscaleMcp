import { startServer } from "./server.js";
import { logger } from "./logger.js";
import { fileURLToPath } from "node:url";

import "dotenv/config";

async function main() {
  try {
    logger.info("Starting Tailscale MCP Server...");
    await startServer();
  } catch (error) {
    logger.error("Failed to start server:", error);
    process.exit(1);
  }
}

process.on("unhandledRejection", (err) => {
  logger.error("Unhandled rejection:", err);
  process.exit(1);
});

// Handle graceful shutdown
process.on("SIGINT", () => {
  logger.info("Received SIGINT, shutting down gracefully...");
  process.exit(0);
});

process.on("SIGTERM", () => {
  logger.info("Received SIGTERM, shutting down gracefully...");
  process.exit(0);
});
// Check if this file is being run directly (works in both ESM and CJS)
const isMainModule = (() => {
  try {
    // ESM check
    if (import.meta.url) {
      return process.argv[1] === fileURLToPath(import.meta.url);
    }
  } catch {
    // CJS fallback - this will be handled by the build process
    return true; // In CJS build, we always want to run main
  }
  return false;
})();

if (isMainModule) {
  main().catch((error) => {
    logger.error("Unhandled error:", error);
    process.exit(1);
  });
}
