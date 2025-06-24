package logger_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/elupevg/golab/logger"
)

func TestLoggerSuccess(t *testing.T) {
	t.Parallel()
	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	log := logger.New(outBuf, errBuf)
	log.Success("test operation")
	got := errBuf.String()
	if got != "" {
		t.Fatalf("errBuf: want \"\", got %q", got)
	}
	want := "[\x1b[32mSUCCESS\x1b[0m] test operation\n"
	got = outBuf.String()
	if want != got {
		t.Errorf("outBuf: want %q, got %q", want, got)
	}
}

func TestLoggerSkipped(t *testing.T) {
	t.Parallel()
	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	log := logger.New(outBuf, errBuf)
	log.Skipped("test operation")
	got := errBuf.String()
	if got != "" {
		t.Fatalf("errBuf: want \"\", got %q", got)
	}
	want := "[\x1b[36mSKIPPED\x1b[0m] test operation\n"
	got = outBuf.String()
	if want != got {
		t.Errorf("outBuf: want %q, got %q", want, got)
	}
}

func TestLoggerErrored(t *testing.T) {
	t.Parallel()
	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	log := logger.New(outBuf, errBuf)
	log.Errored(errors.New("test error"))
	got := outBuf.String()
	if got != "" {
		t.Fatalf("outBuf: want \"\", got %q", got)
	}
	want := "[\x1b[31mERROR\x1b[0m] test error\n"
	got = errBuf.String()
	if want != got {
		t.Errorf("errBuf: want %q, got %q", want, got)
	}
}
