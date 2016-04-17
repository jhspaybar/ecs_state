package ecsstate

import (
	"log"
	"os"
)

// Wrap a logger of your choice with this interface to allow the ecs_state.State object to log.
type logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

// Allows us to create a default logger if the user does not provide on.
type Logger struct {
	*log.Logger
}

// A logger that logs on Stdout provided to quickly get started.
var DefaultLogger = Logger{log.New(os.Stdout, "\r\n", 0)}

// Log at a debug level
func (logger *Logger) Debug(args ...interface{}) {
	logger.Println("[DEBUG]", args)
}

// Log at an info level
func (logger *Logger) Info(args ...interface{}) {
	logger.Println("[INFO]", args)
}

// Log at a warn level
func (logger *Logger) Warn(args ...interface{}) {
	logger.Println("[WARN]", args)
}

// Log at an error level
func (logger *Logger) Error(args ...interface{}) {
	logger.Println("[ERROR]", args)
}
