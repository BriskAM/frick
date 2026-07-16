package shell

import (
	"strings"
	"testing"
)

func TestGetIntegrationScript(t *testing.T) {
	script := GetIntegrationScript()
	if len(script) == 0 {
		t.Fatal("Expected integration script to be non-empty")
	}

	if !strings.Contains(script, "$ZSH_VERSION") {
		t.Error("Expected script to mention Zsh version check")
	}
	if !strings.Contains(script, "$BASH_VERSION") {
		t.Error("Expected script to mention Bash version check")
	}
	if !strings.Contains(script, "command frick") {
		t.Error("Expected script to forward commands to binary")
	}
}
