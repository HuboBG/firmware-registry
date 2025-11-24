package logging

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"

	"firmware-registry-api/internal/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Setup initializes the global logger based on configuration
func Setup(cfg config.Config) error {
	// Set log level
	level, err := parseLevel(cfg.Logging.Level)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", cfg.Logging.Level, err)
	}
	zerolog.SetGlobalLevel(level)

	// Set time format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Create writer based on output configuration
	var writer io.Writer
	switch strings.ToLower(cfg.Logging.Output) {
	case "stdout":
		writer = setupConsoleWriter(cfg)
	case "file":
		writer, err = setupFileWriter(cfg)
		if err != nil {
			return fmt.Errorf("failed to setup file writer: %w", err)
		}
	case "syslog":
		writer, err = setupSyslogWriter(cfg)
		if err != nil {
			return fmt.Errorf("failed to setup syslog writer: %w", err)
		}
	case "multi":
		writer, err = setupMultiWriter(cfg)
		if err != nil {
			return fmt.Errorf("failed to setup multi writer: %w", err)
		}
	default:
		return fmt.Errorf("invalid log output %q", cfg.Logging.Output)
	}

	// Set global logger
	log.Logger = zerolog.New(writer).With().Timestamp().Caller().Logger()

	log.Info().
		Str("level", cfg.Logging.Level).
		Str("format", cfg.Logging.Format).
		Str("output", cfg.Logging.Output).
		Msg("Logger initialized")

	return nil
}

func parseLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel, nil
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	case "disabled":
		return zerolog.Disabled, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown level: %s", level)
	}
}

func setupConsoleWriter(cfg config.Config) io.Writer {
	if strings.ToLower(cfg.Logging.Format) == "console" {
		// Pretty console output for development
		return zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}
	// JSON output to stdout
	return os.Stdout
}

func setupFileWriter(cfg config.Config) (io.Writer, error) {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(cfg.Logging.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Setup log rotation with lumberjack
	writer := &lumberjack.Logger{
		Filename:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSizeMB,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAgeDays,
		Compress:   cfg.Logging.Compress,
		LocalTime:  true,
	}

	return writer, nil
}

func setupSyslogWriter(cfg config.Config) (io.Writer, error) {
	var writer *syslog.Writer
	var err error

	// Determine syslog connection type
	if cfg.Logging.SyslogAddr == "" {
		// Local syslog
		writer, err = syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "firmware-registry")
	} else {
		// Remote syslog
		network := cfg.Logging.SyslogNet
		if network == "" {
			network = "udp"
		}
		writer, err = syslog.Dial(network, cfg.Logging.SyslogAddr, syslog.LOG_INFO|syslog.LOG_DAEMON, "firmware-registry")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}

	return writer, nil
}

func setupMultiWriter(cfg config.Config) (io.Writer, error) {
	var writers []io.Writer

	// Always include stdout/console
	writers = append(writers, setupConsoleWriter(cfg))

	// Add file writer if path is configured
	if cfg.Logging.FilePath != "" {
		fileWriter, err := setupFileWriter(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to setup file writer: %w", err)
		}
		writers = append(writers, fileWriter)
	}

	// Add syslog writer if address is configured
	if cfg.Logging.SyslogAddr != "" {
		syslogWriter, err := setupSyslogWriter(cfg)
		if err != nil {
			// Log warning but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to setup syslog writer: %v\n", err)
		} else {
			writers = append(writers, syslogWriter)
		}
	}

	return zerolog.MultiLevelWriter(writers...), nil
}

// GetLogger returns the global logger
func GetLogger() *zerolog.Logger {
	return &log.Logger
}
