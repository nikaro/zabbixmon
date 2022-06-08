package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog"
)

func TestSetLogLevelPanic(t *testing.T) {
	setLogLevel("panic")
	if zerolog.GlobalLevel() != zerolog.PanicLevel {
		t.Errorf("got: %v, want: %v", zerolog.GlobalLevel(), zerolog.PanicLevel)
	}
}

func TestSetLogLevelWarn(t *testing.T) {
	setLogLevel("warn")
	if zerolog.GlobalLevel() != zerolog.WarnLevel {
		t.Errorf("got: %v, want: %v", zerolog.GlobalLevel(), zerolog.WarnLevel)
	}
}

func TestSetLogLevelNotFound(t *testing.T) {
	if os.Getenv("LOG_NOT_FOUND") == "1" {
		setLogLevel("not_found")
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestSetLogLevelNotFound")
	cmd.Env = append(os.Environ(), "LOG_NOT_FOUND=1")
	if err := cmd.Run(); err == nil {
		t.Errorf("got: %v, want exit status 1", err)
	}
}
