package logger

import (
	"log/slog"
	"os"
)

// An implementation of the logger.Logger.
// Based on the slog in the standard library.
type SlogLogger struct {
	logger *slog.Logger
}

// Creates a new instance of the SlogLogger
//
// Returns:
//   - a pointer to a SlogLogger that logs in JSON format
func NewSlogLogger() *SlogLogger {
	return &SlogLogger{logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))}
}

// Logs an info message
//
// Parameters:
//   - msg: a human friendly message
//   - args: alternating keys and values (ex: "event", "eventName", "user", "userId")
func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Logs a warn message
//
// Parameters:
//   - msg: a human friendly message
//   - args: alternating keys and values (ex: "event", "eventName", "user", "userId")
func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Logs an error message
//
// Parameters:
//   - msg: a human friendly message
//   - args: alternating keys and values (ex: "event", "eventName", "user", "userId")
func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Logs an error message and exits.
// Similar to the standard library's log.Fatal().
// This function calls os.Exit(1) after logging the error
//
// Parameters:
//   - msg: a human friendly message
//   - args: alternating keys and values (ex: "event", "eventName", "user", "userId")
func (l *SlogLogger) Fatal(msg string, args ...any) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}
