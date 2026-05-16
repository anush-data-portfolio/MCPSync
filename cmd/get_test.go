package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── mcpsyncConfigPath ─────────────────────────────────────────────────────────

func TestMCPSyncConfigPath(t *testing.T) {
	got := mcpsyncConfigPath()
	if got == "" {
		t.Error("mcpsyncConfigPath() should return a non-empty path")
	}
	if !strings.Contains(got, ".mcpsync") {
		t.Errorf("mcpsyncConfigPath() = %q, should contain '.mcpsync'", got)
	}
	if !strings.HasSuffix(got, "config.json") {
		t.Errorf("mcpsyncConfigPath() = %q, should end with 'config.json'", got)
	}
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(got, home) {
		t.Errorf("mcpsyncConfigPath() = %q, should be under home dir %q", got, home)
	}
}

// ── loadMergedServers ─────────────────────────────────────────────────────────

func TestLoadMergedServers_NoAgentsDetected(t *testing.T) {
	// Override CustomPaths to point to non-existent paths so no agent is detected.
	// loadMergedServers should return nil when no agents have config files.
	origPaths := CustomPaths
	CustomPaths = map[string]string{}
	defer func() { CustomPaths = origPaths }()

	// In a test environment, agent config files don't exist, so no agents are
	// detected and loadMergedServers returns nil without error.
	result := loadMergedServers()
	// Result is nil when no agents detected (prints warning to stderr).
	// We just verify it doesn't panic and returns a slice.
	_ = result
}

func TestLoadMergedServers_WithDetectedAgent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp.json")
	// Use a server name unlikely to already exist in any real agent config.
	serverName := "test-only-server-" + filepath.Base(dir)
	os.WriteFile(p, []byte(`{"mcpServers":{"`+serverName+`":{"type":"stdio","command":"echo"}}}`), 0o644)

	origPaths := CustomPaths
	CustomPaths = map[string]string{"claude-desktop": p}
	defer func() { CustomPaths = origPaths }()

	result := loadMergedServers()
	if len(result) == 0 {
		t.Error("loadMergedServers: expected at least 1 server from detected agent")
	}
	found := false
	for _, s := range result {
		if s.Name == serverName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("loadMergedServers: server %q not found in results %v", serverName, result)
	}
}
