import { writeFileSync, appendFileSync } from "fs";

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

class Logger {
  private level: LogLevel;
  private logFilePath: string | null = null;

  constructor(level: LogLevel = LogLevel.INFO) {
    this.level = level;

    // Initialize file logging if environment variable is set
    if (process.env.MCP_SERVER_LOG_FILE) {
      const timestamp = new Date().toISOString().replace(/[:.]/g, "-");
      this.logFilePath = process.env.MCP_SERVER_LOG_FILE.replace(
        "{timestamp}",
        timestamp
      );

      // Create initial log file with header
      const header = `=== Tailscale MCP Server Log ===\nStarted: ${new Date().toISOString()}\nLog Level: ${
        LogLevel[level]
      }\n\n`;
      try {
        writeFileSync(this.logFilePath, header, "utf8");
        console.info(`📝 Server logging to file: ${this.logFilePath}`);
      } catch (error) {
        console.error(`❌ Failed to create server log file: ${error}`);
        this.logFilePath = null;
      }
    }
  }

  setLevel(level: LogLevel): void {
    this.level = level;
    if (this.logFilePath) {
      this.writeToFile(`Log level changed to: ${LogLevel[level]}`);
    }
  }

  private writeToFile(message: string): void {
    if (this.logFilePath) {
      try {
        appendFileSync(this.logFilePath, message + "\n", "utf8");
      } catch (error) {
        console.error(`❌ Failed to write to server log file: ${error}`);
      }
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

      // Write to file first (without console formatting)
      if (this.logFilePath) {
        this.writeToFile(`${prefix} ${fullMessage}`);
      }

      // Then write to console
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
}

// Export singleton instance
export const logger = new Logger(
  process.env.LOG_LEVEL ? parseInt(process.env.LOG_LEVEL) : LogLevel.INFO
);

// Export class for custom instances
export { Logger };
