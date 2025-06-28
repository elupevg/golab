// Package logger provides means to print colorized log messages on the screen.
package logger

import (
	"fmt"
	"io"
)

const (
	reset = "\x1b[0m"
	red   = "\x1b[31m"
	green = "\x1b[32m"
	cyan  = "\x1b[36m"
)

// Logger implements a simple logger with customizable out and error writers.
type Logger struct {
	out io.Writer
	err io.Writer
}

// New creates and returns a new Logger instance.
func New(out, err io.Writer) *Logger {
	return &Logger{
		out: out,
		err: err,
	}
}

// Success annotates the provided message with colorized prefix and prints it.
func (l *Logger) Success(msg string) {
	fmt.Fprintf(l.out, "[%sSUCCESS%s] %s\n", green, reset, msg)
}

// Skipped annotates the provided message with colorized prefix and prints it.
func (l *Logger) Skipped(msg string) {
	fmt.Fprintf(l.out, "[%sSKIPPED%s] %s\n", cyan, reset, msg)
}

// Errored annotates the provided error message with colorized prefix and prints it.
func (l *Logger) Errored(err error) {
	fmt.Fprintf(l.err, "[%sERROR%s] %s\n", red, reset, err.Error())
}
