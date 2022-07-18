package log

import (
	"fmt"
	"github.com/cdfmlr/crud/pkg/ginlogrus"
	"github.com/cdfmlr/crud/pkg/gormlogrus"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Level is the level of log: LevelDebug, LevelInfo, LevelWarn, LevelError
type Level string

// Log levels.
const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Logger is the global Logger instance
var Logger = logrus.New()

// Logger4Gorm is the wrapped Logger instance for Gorm
var Logger4Gorm *gormlogrus.Logger

// Logger4Gin is a gin middleware for Logger
var Logger4Gin gin.HandlerFunc

// UseLogger use given logger instance to initializes global Logger.
// The Logger instance will be shared by the whole crud package
// (and the underlying GORM, Gin included)
func UseLogger(logger *logrus.Logger, options ...LoggerOption) {
	Logger = logger

	for _, option := range options {
		option(logger)
	}

	Logger4Gorm = gormlogrus.Use(Logger)
	Logger4Gin = ginlogrus.Logger(Logger)
}

// LoggerOption is a function that can be used to configure the global Logger
type LoggerOption func(logger *logrus.Logger)

// WithLevel sets the Logger level.
//
// The level can be one of the following:
// LevelDebug, LevelInfo, LevelWarn, LevelError.
// If the level is not valid, it will use Debug by default.
//
// Note: this option will affect the log level of given logger
// instance when it is used by UseLogger(logger, WithLevel(...)).
func WithLevel(level Level) LoggerOption {
	return func(logger *logrus.Logger) {
		logger.SetLevel(getLogrusLevel(level))
	}
}

// Level -> logrus.Level
func getLogrusLevel(level Level) logrus.Level {
	switch level {
	case LevelDebug:
		return logrus.DebugLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelError:
		return logrus.ErrorLevel
	default:
		fmt.Printf("getLogrusLevel: Unknown Logger level: %s, using Debug by default", level)
		return logrus.DebugLevel
	}
}

// NewLogger creates and use a new logrus.Logger instance
func NewLogger(options ...LoggerOption) *logrus.Logger {
	UseLogger(logrus.New(), options...)
	return Logger
}

// By default, creates a new logger with LevelDebug,
// this will be overridden by UseLogger
func init() {
	NewLogger(WithLevel(LevelDebug))
}
