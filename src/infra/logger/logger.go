package logger

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// Initialize sets up the global logger
func Initialize(isDevelopment bool) error {
	var config zap.Config
	
	if isDevelopment {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var err error
	Logger, err = config.Build()
	if err != nil {
		return err
	}

	Sugar = Logger.Sugar()
	return nil
}

// Close flushes the logger
func Close() {
	if Logger != nil {
		Logger.Sync()
	}
}

// CorrelationIDKey is the context key for correlation ID
const CorrelationIDKey = "correlation_id"

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context) context.Context {
	correlationID := uuid.New().String()
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID gets correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// LoggerMiddleware is a Fiber middleware for structured logging
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		// Generate correlation ID
		correlationID := uuid.New().String()
		c.Locals(CorrelationIDKey, correlationID)
		c.Set("X-Correlation-ID", correlationID)

		// Process request
		err := c.Next()

		// Log request
		fields := []zap.Field{
			zap.String("correlation_id", correlationID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.Int("response_size", len(c.Response().Body())),
		}

		// Add user info if available
		if user := c.Locals("user"); user != nil {
			if claims, ok := user.(map[string]interface{}); ok {
				if userID, exists := claims["user_id"].(string); exists {
					fields = append(fields, zap.String("user_id", userID))
				}
				if role, exists := claims["role"].(string); exists {
					fields = append(fields, zap.String("user_role", role))
				}
			}
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			Logger.Error("Request failed", fields...)
		} else {
			Logger.Info("Request completed", fields...)
		}

		return err
	}
}

// LogWithCorrelation logs with correlation ID from context
func LogWithCorrelation(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}
	
	switch level {
	case zapcore.DebugLevel:
		Logger.Debug(msg, fields...)
	case zapcore.InfoLevel:
		Logger.Info(msg, fields...)
	case zapcore.WarnLevel:
		Logger.Warn(msg, fields...)
	case zapcore.ErrorLevel:
		Logger.Error(msg, fields...)
	case zapcore.FatalLevel:
		Logger.Fatal(msg, fields...)
	}
}

// Helper functions for common logging patterns
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	LogWithCorrelation(ctx, zapcore.InfoLevel, msg, fields...)
}

func Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	LogWithCorrelation(ctx, zapcore.ErrorLevel, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	LogWithCorrelation(ctx, zapcore.WarnLevel, msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	LogWithCorrelation(ctx, zapcore.DebugLevel, msg, fields...)
}