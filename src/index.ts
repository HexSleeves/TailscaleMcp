import "dotenv/config";
import { startServer } from "./server.js";
import { logger } from "./logger.js";
import { fileURLToPath } from "node:url";

async function main() {
  try {
    logger.info("Starting Tailscale MCP Server...");
    await startServer();
  } catch (error) {
    logger.error("Failed to start server:", error);
    await gracefulShutdown(1);
  }
}

async function gracefulShutdown(exitCode: number = 0): Promise<void> {
  try {
    logger.info("Flushing logs before shutdown...");
    await logger.flush();
    await logger.close();
  } catch (error) {
    console.error("Error during graceful shutdown:", error);
  } finally {
    process.exit(exitCode);
  }
}

process.on("unhandledRejection", async (err) => {
  logger.error("Unhandled rejection:", err);
  await gracefulShutdown(1);
});

// Handle graceful shutdown
process.on("SIGINT", async () => {
  logger.info("Received SIGINT, shutting down gracefully...");
  await gracefulShutdown(0);
});

process.on("SIGTERM", async () => {
  logger.info("Received SIGTERM, shutting down gracefully...");
  await gracefulShutdown(0);
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
  main().catch(async (error) => {
    logger.error("Unhandled error:", error);
    await gracefulShutdown(1);
  });
}
