// Simple logger for the UI that logs to console
// In production, errors can also be sent to a backend logging service

export enum LogLevel {
    DEBUG = 'debug',
    INFO = 'info',
    WARN = 'warn',
    ERROR = 'error',
}

interface LogEntry {
    timestamp: string;
    level: LogLevel;
    message: string;
    context?: Record<string, any>;
    error?: Error;
}

class Logger {
    private minLevel: LogLevel;

    constructor() {
        // Set minimum log level based on environment
        this.minLevel = import.meta.env.DEV ? LogLevel.DEBUG : LogLevel.INFO;
    }

    private shouldLog(level: LogLevel): boolean {
        const levels = [LogLevel.DEBUG, LogLevel.INFO, LogLevel.WARN, LogLevel.ERROR];
        const currentLevelIndex = levels.indexOf(this.minLevel);
        const messageLevelIndex = levels.indexOf(level);
        return messageLevelIndex >= currentLevelIndex;
    }

    private formatLog(entry: LogEntry): void {
        const { timestamp, level, message, context, error } = entry;

        // Format for console
        const style = this.getStyle(level);
        const prefix = `[${timestamp}] [${level.toUpperCase()}]`;

        if (error) {
            console.error(prefix, message, context || '', error);
        } else {
            switch (level) {
                case LogLevel.ERROR:
                    console.error(prefix, message, context || '');
                    break;
                case LogLevel.WARN:
                    console.warn(prefix, message, context || '');
                    break;
                case LogLevel.INFO:
                    console.info(prefix, message, context || '');
                    break;
                case LogLevel.DEBUG:
                default:
                    console.log(prefix, message, context || '');
                    break;
            }
        }

        // In production, you could send errors to a backend logging service
        // if (level === LogLevel.ERROR && !import.meta.env.DEV) {
        //     this.sendToBackend(entry);
        // }
    }

    private getStyle(level: LogLevel): string {
        switch (level) {
            case LogLevel.ERROR:
                return 'color: red; font-weight: bold;';
            case LogLevel.WARN:
                return 'color: orange; font-weight: bold;';
            case LogLevel.INFO:
                return 'color: blue;';
            case LogLevel.DEBUG:
            default:
                return 'color: gray;';
        }
    }

    private log(level: LogLevel, message: string, context?: Record<string, any>, error?: Error): void {
        if (!this.shouldLog(level)) {
            return;
        }

        const entry: LogEntry = {
            timestamp: new Date().toISOString(),
            level,
            message,
            context,
            error,
        };

        this.formatLog(entry);
    }

    debug(message: string, context?: Record<string, any>): void {
        this.log(LogLevel.DEBUG, message, context);
    }

    info(message: string, context?: Record<string, any>): void {
        this.log(LogLevel.INFO, message, context);
    }

    warn(message: string, context?: Record<string, any>): void {
        this.log(LogLevel.WARN, message, context);
    }

    error(message: string, error?: Error, context?: Record<string, any>): void {
        this.log(LogLevel.ERROR, message, context, error);
    }
}

// Export singleton instance
export const logger = new Logger();
