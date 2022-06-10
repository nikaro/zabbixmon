package config

import (
	"os"
	"os/exec"
	"testing"
)

func TestInitConfigNotFound(t *testing.T) {
	if os.Getenv("CONF_NOT_FOUND") == "1" {
		InitConfig()
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestInitConfigNotFound")
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	cmd.Env = append(os.Environ(), "CONF_NOT_FOUND=1")
	if err := cmd.Run(); err == nil {
		t.Errorf("got: %v, want exit status 1", err)
	}
}
