package logging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart2405/API_Gateway/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger initializes the structured logger
func InitLogger() error {
	config := config.GetConfig()
	if config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	var zapConfig zap.Config

	switch config.Logging.Level {
	case "debug":
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Configure output
	if config.Logging.OutputPath != "stdout" {
		zapConfig.OutputPaths = []string{config.Logging.OutputPath}
		zapConfig.ErrorOutputPaths = []string{config.Logging.OutputPath}
	}

	// Configure encoding
	if config.Logging.Format == "json" {
		zapConfig.Encoding = "json"
	} else {
		zapConfig.Encoding = "console"
	}

	// Configure time format
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var err error
	Logger, err = zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Replace global logger
	zap.ReplaceGlobals(Logger)

	log.Printf("Logger initialized with level: %s, format: %s", config.Logging.Level, config.Logging.Format)
	return nil
}

// GetLogger returns the logger instance
func GetLogger() *zap.Logger {
	return Logger
}

// LoggingMiddleware provides structured logging for HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Create structured log entry
		logger := Logger.With(
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("ip", param.ClientIP),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.String("error", param.ErrorMessage),
		)

		// Log based on status code
		switch {
		case param.StatusCode >= 500:
			logger.Error("Server error")
		case param.StatusCode >= 400:
			logger.Warn("Client error")
		default:
			logger.Info("Request completed")
		}

		return ""
	})
}

// RequestLogger logs detailed request information
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Add request ID to context
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}

		// Create logger with request context
		logger := Logger.With(
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Add logger to context
		c.Set("logger", logger)

		// Process request
		c.Next()

		// Log response
		duration := time.Since(start)
		logger = logger.With(
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("duration", duration),
		)

		// Log errors
		if len(c.Errors) > 0 {
			errorMsgs := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errorMsgs[i] = err.Error()
			}
			logger.Error("Request completed with errors", zap.Strings("errors", errorMsgs))
		} else {
			logger.Info("Request completed successfully")
		}
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// GetLoggerFromContext gets logger from gin context
func GetLoggerFromContext(c *gin.Context) *zap.Logger {
	if logger, exists := c.Get("logger"); exists {
		if zapLogger, ok := logger.(*zap.Logger); ok {
			return zapLogger
		}
	}
	return Logger
}

// LogError logs an error with context
func LogError(ctx context.Context, msg string, err error, fields ...zap.Field) {
	logger := Logger
	if ctx != nil {
		if ginCtx, ok := ctx.(*gin.Context); ok {
			logger = GetLoggerFromContext(ginCtx)
		}
	}

	allFields := append(fields, zap.Error(err))
	logger.Error(msg, allFields...)
}

// LogInfo logs info with context
func LogInfo(ctx context.Context, msg string, fields ...zap.Field) {
	logger := Logger
	if ctx != nil {
		if ginCtx, ok := ctx.(*gin.Context); ok {
			logger = GetLoggerFromContext(ginCtx)
		}
	}

	logger.Info(msg, fields...)
}

// LogDebug logs debug with context
func LogDebug(ctx context.Context, msg string, fields ...zap.Field) {
	logger := Logger
	if ctx != nil {
		if ginCtx, ok := ctx.(*gin.Context); ok {
			logger = GetLoggerFromContext(ginCtx)
		}
	}

	logger.Debug(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}
