package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger instance.
var (
	Log  *zap.Logger //store the logger for entire application
	once sync.Once   //ensure the logger is initialized once
)

// Init initializes the global logger based on the application mode.
// mode: "debug" for development (console, debug level),
//
//	"release" or "test" for production (JSON, info level).
func Init(mode string) {
	once.Do(func() {
		var config zap.Config

		if mode == "debug" {
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		} else {
			config = zap.NewProductionConfig() //Logs are JSON
			config.EncoderConfig.TimeKey = "timestamp"
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		}

		var err error
		Log, err = config.Build( //creates the logger
			zap.AddCaller(), //Add the file and line number
			zap.AddCallerSkip(0),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		if err != nil {
			// Fallback to a basic logger if the configured one fails.
			Log = zap.NewExample()
			Log.Error("failed to initialize configured logger, using fallback", zap.Error(err))
		}
	})
}

// Sync flushes any buffered log entries. Should be called on shutdown.
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// WithRequestID returns a new logger with the request_id field attached.
func WithRequestID(requestID string) *zap.Logger {
	return Log.With(zap.String("request_id", requestID))
}

// WithUserID returns a new logger with the user_id field attached.
func WithUserID(userID string) *zap.Logger {
	return Log.With(zap.String("user_id", userID))
}

// WithFields returns a new logger with multiple fields attached.
func WithFields(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// Fatal logs a fatal message and exits the process.
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
	os.Exit(1)
}

//Single Design Pattern
