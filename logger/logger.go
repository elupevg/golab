// Package logger provides means to print colorized log messages on the screen.
package logger

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Logger implements a simple logger with customizable out and error writers.
type Logger struct {
	out   io.Writer
	err   io.Writer
	green func(a ...interface{}) string
	cyan  func(a ...interface{}) string
	red   func(a ...interface{}) string
}

// New creates and returns a new Logger instance.
func New(out, err io.Writer) *Logger {
	return &Logger{
		out:   out,
		err:   err,
		green: color.New(color.FgGreen).SprintFunc(),
		cyan:  color.New(color.FgCyan).SprintFunc(),
		red:   color.New(color.FgRed).SprintFunc(),
	}
}

// Success annotates the provided message with colorized prefix and prints it.
func (l *Logger) Success(msg string) {
	fmt.Fprintf(l.out, "[%s] %s\n", l.green("SUCCESS"), msg)
}

// Skipped annotates the provided message with colorized prefix and prints it.
func (l *Logger) Skipped(msg string) {
	fmt.Fprintf(l.out, "[%s] %s\n", l.cyan("SKIPPED"), msg)
}

// Errored annotates the provided error message with colorized prefix and prints it.
func (l *Logger) Errored(err error) {
	fmt.Fprintf(l.err, "[%s] %s\n", l.red("ERROR"), err.Error())
}
