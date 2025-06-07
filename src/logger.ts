import { writeFile, appendFile } from "fs/promises";

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

interface LogEntry {
  message: string;
  resolve: () => void;
  reject: (error: Error) => void;
}

class Logger {
  private level: LogLevel;
  private logFilePath: string | null = null;
  private writeQueue: LogEntry[] = [];
  private isProcessingQueue = false;
  private readonly maxBatchSize = 50; // Process up to 50 entries at once
  private readonly flushTimeout = 100; // Flush every 100ms if queue has items
  private flushTimer: NodeJS.Timeout | null = null;

  constructor(level: LogLevel = LogLevel.INFO) {
    this.level = level;

    // Initialize file logging if environment variable is set
    if (process.env.MCP_SERVER_LOG_FILE) {
      let logPath = process.env.MCP_SERVER_LOG_FILE;
      if (!logPath) {
        console.warn(
          "Warning: MCP_SERVER_LOG_FILE is not set or is empty. File logging will be disabled."
        );
        return;
      }

      // If the log path contains {timestamp}, replace it with the timestamp
      if (logPath.includes("{timestamp}")) {
        const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
        logPath = logPath.replace("{timestamp}", timestamp);
      }

      this.logFilePath = logPath;

      // Create initial log file with header asynchronously
      this.initializeLogFile(level);
    }
  }

  private async initializeLogFile(level: LogLevel): Promise<void> {
    if (!this.logFilePath) return;

    const header = `=== Tailscale MCP Server Log ===\nStarted: ${new Date().toISOString()}\nLog Level: ${
      LogLevel[level]
    }\n\n`;

    try {
      await writeFile(this.logFilePath, header, "utf8");
      console.debug(`📝 Server logging to file: ${this.logFilePath}`);
    } catch (error) {
      console.error(`❌ Failed to create server log file: ${error}`);
      this.logFilePath = null;
    }
  }

  setLevel(level: LogLevel): void {
    this.level = level;
    if (this.logFilePath) {
      // Don't await this to avoid blocking
      this.writeToFileAsync(`Log level changed to: ${LogLevel[level]}`).catch(
        (error) => {
          console.error(`❌ Failed to log level change: ${error}`);
        }
      );
    }
  }

  private async writeToFileAsync(message: string): Promise<void> {
    if (!this.logFilePath) return;

    return new Promise((resolve, reject) => {
      this.writeQueue.push({
        message: message + "\n",
        resolve,
        reject,
      });

      // Schedule processing if not already processing
      this.scheduleQueueProcessing();
    });
  }

  private scheduleQueueProcessing(): void {
    if (this.isProcessingQueue) return;

    // Process immediately if queue is large enough
    if (this.writeQueue.length >= this.maxBatchSize) {
      this.processWriteQueue();
      return;
    }

    // Otherwise, schedule processing with a timeout
    if (!this.flushTimer) {
      this.flushTimer = setTimeout(() => {
        this.flushTimer = null;
        this.processWriteQueue();
      }, this.flushTimeout);
    }
  }

  private async processWriteQueue(): Promise<void> {
    if (this.isProcessingQueue || this.writeQueue.length === 0) return;

    this.isProcessingQueue = true;

    // Clear the flush timer since we're processing now
    if (this.flushTimer) {
      clearTimeout(this.flushTimer);
      this.flushTimer = null;
    }

    // Process entries in batches
    const batch: LogEntry[] = [];
    let batchMessages = "";

    // Collect a batch of entries
    while (this.writeQueue.length > 0 && batch.length < this.maxBatchSize) {
      const entry = this.writeQueue.shift();
      if (entry) {
        batch.push(entry);
        batchMessages += entry.message;
      }
    }

    if (batch.length > 0 && this.logFilePath) {
      try {
        // Write all messages in the batch at once
        await appendFile(this.logFilePath, batchMessages, "utf8");

        // Resolve all promises in the batch
        batch.forEach((entry) => entry.resolve());
      } catch (error) {
        console.error(`❌ Failed to write batch to server log file: ${error}`);

        // Reject all promises in the batch
        batch.forEach((entry) => entry.reject(error as Error));
      }
    }

    this.isProcessingQueue = false;

    // If there are more items in the queue, schedule another processing round
    if (this.writeQueue.length > 0) {
      this.scheduleQueueProcessing();
    }
  }

  private log(level: LogLevel, message: string, ...args: any[]): void {
    if (level >= this.level) {
      const timestamp = new Date().toISOString();
      const levelName = LogLevel[level];
      const prefix = `[${timestamp}] [${levelName}]`;
      const fullMessage =
        args.length > 0
          ? `${message} ${args
              .map((arg) =>
                typeof arg === "object" ? JSON.stringify(arg) : String(arg)
              )
              .join(" ")}`
          : message;

      // Write to file asynchronously (non-blocking)
      if (this.logFilePath) {
        this.writeToFileAsync(`${prefix} ${fullMessage}`).catch((error) => {
          // Silently handle file write errors to avoid infinite loops
          console.error(`❌ Async log write failed: ${error.message}`);
        });
      }

      // Write to console immediately
      switch (level) {
        case LogLevel.DEBUG:
          console.debug(prefix, message, ...args);
          break;
        case LogLevel.INFO:
          console.info(prefix, message, ...args);
          break;
        case LogLevel.WARN:
          console.warn(prefix, message, ...args);
          break;
        case LogLevel.ERROR:
          console.error(prefix, message, ...args);
          break;
      }
    }
  }

  debug(message: string, ...args: any[]): void {
    this.log(LogLevel.DEBUG, message, ...args);
  }

  info(message: string, ...args: any[]): void {
    this.log(LogLevel.INFO, message, ...args);
  }

  warn(message: string, ...args: any[]): void {
    this.log(LogLevel.WARN, message, ...args);
  }

  error(message: string, ...args: any[]): void {
    this.log(LogLevel.ERROR, message, ...args);
  }

  // Helper method for structured logging
  logObject(level: LogLevel, message: string, obj: any): void {
    this.log(level, message, JSON.stringify(obj, null, 2));
  }

  // Method to flush pending writes (useful for graceful shutdown)
  async flush(): Promise<void> {
    // Clear any pending flush timer
    if (this.flushTimer) {
      clearTimeout(this.flushTimer);
      this.flushTimer = null;
    }

    // Process any remaining items in the queue
    while (this.writeQueue.length > 0) {
      await this.processWriteQueue();
    }
  }

  // Method to close the logger and flush remaining writes
  async close(): Promise<void> {
    await this.flush();
    this.logFilePath = null;
  }
}

// Export singleton instance
export const logger = new Logger(
  process.env.LOG_LEVEL ? parseInt(process.env.LOG_LEVEL) : LogLevel.INFO
);

// Export class for custom instances
export { Logger };
