package cli_test

import (
	"testing"

	"archvpn/internal/cli"
)

func TestCLIHelp(t *testing.T) {
	err := cli.Run([]string{"help"})
	if err != nil {
		t.Fatalf("cli.Run(help) failed: %v", err)
	}
}

func TestCLIUnknownCommand(t *testing.T) {
	err := cli.Run([]string{"nonexistent-command-xyz"})
	if err == nil {
		t.Errorf("expected error for unknown command, got nil")
	}
}
