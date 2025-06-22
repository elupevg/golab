package logger

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Logger implements a simple logger to print colored log messages.
type Logger struct {
	out, err         io.Writer
	green, cyan, red func(a ...interface{}) string
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

// Success annotates the provided message with colorized prefix and dumps it to the Logger's writer.
func (l *Logger) Success(msg string) {
	fmt.Fprintf(l.out, "[%s] %s\n", l.green("SUCCESS"), msg)
}

// Skipped annotates the provided message with colorized prefix and dumps it to the Logger's writer.
func (l *Logger) Skipped(msg string) {
	fmt.Fprintf(l.out, "[%s] %s\n", l.cyan("SKIPPED"), msg)
}

// Skipped annotates the provided message with colorized prefix and dumps it to the Logger's writer.
func (l *Logger) Error(err error) {
	fmt.Fprintf(l.err, "[%s] %s\n", l.red("ERROR"), err.Error())
}
