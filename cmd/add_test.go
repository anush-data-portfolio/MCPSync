package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ── autoDetectKey ─────────────────────────────────────────────────────────────

func TestAutoDetectKey_MCPServers(t *testing.T) {
	raw := map[string]any{"mcpServers": map[string]any{}}
	if got := autoDetectKey(raw); got != "mcpServers" {
		t.Errorf("autoDetectKey: got %q, want 'mcpServers'", got)
	}
}

func TestAutoDetectKey_Servers(t *testing.T) {
	raw := map[string]any{"servers": map[string]any{}}
	if got := autoDetectKey(raw); got != "servers" {
		t.Errorf("autoDetectKey: got %q, want 'servers'", got)
	}
}

func TestAutoDetectKey_ContextServers(t *testing.T) {
	raw := map[string]any{"context_servers": map[string]any{}}
	if got := autoDetectKey(raw); got != "context_servers" {
		t.Errorf("autoDetectKey: got %q, want 'context_servers'", got)
	}
}

func TestAutoDetectKey_PrioritizesMCPServers(t *testing.T) {
	// mcpServers should win if both are present
	raw := map[string]any{
		"mcpServers": map[string]any{},
		"servers":    map[string]any{},
	}
	if got := autoDetectKey(raw); got != "mcpServers" {
		t.Errorf("autoDetectKey: got %q, want 'mcpServers' (higher priority)", got)
	}
}

func TestAutoDetectKey_Unknown(t *testing.T) {
	raw := map[string]any{"foo": map[string]any{}}
	if got := autoDetectKey(raw); got != "" {
		t.Errorf("autoDetectKey(unknown): got %q, want ''", got)
	}
}

func TestAutoDetectKey_Empty(t *testing.T) {
	if got := autoDetectKey(map[string]any{}); got != "" {
		t.Errorf("autoDetectKey(empty): got %q, want ''", got)
	}
}

// ── loadMCPSyncConfig / saveMCPSyncConfig ─────────────────────────────────────

func TestLoadMCPSyncConfig_NonExistent(t *testing.T) {
	cfg := loadMCPSyncConfig("/nonexistent/config.json")
	if len(cfg.CustomAgents) != 0 {
		t.Errorf("loadMCPSyncConfig(nonexistent): expected empty config, got %+v", cfg)
	}
}

func TestLoadMCPSyncConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	os.WriteFile(p, []byte(`{not valid`), 0o644)

	cfg := loadMCPSyncConfig(p)
	if len(cfg.CustomAgents) != 0 {
		t.Error("loadMCPSyncConfig(bad JSON): expected empty config")
	}
}

func TestSaveMCPSyncConfig_BasicRoundtrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	cfg := mcpsyncConfig{
		Version: 1,
		CustomAgents: []customAgentEntry{
			{ID: "my-agent", DisplayName: "My Agent", ConfigPath: "/path/to/mcp.json", KeyPath: "mcpServers"},
		},
	}
	if err := saveMCPSyncConfig(p, cfg); err != nil {
		t.Fatalf("saveMCPSyncConfig: %v", err)
	}

	loaded := loadMCPSyncConfig(p)
	if loaded.Version != 1 {
		t.Errorf("Version = %d, want 1", loaded.Version)
	}
	if len(loaded.CustomAgents) != 1 {
		t.Fatalf("expected 1 custom agent, got %d", len(loaded.CustomAgents))
	}
	if loaded.CustomAgents[0].ID != "my-agent" {
		t.Errorf("ID = %q, want 'my-agent'", loaded.CustomAgents[0].ID)
	}
	if loaded.CustomAgents[0].DisplayName != "My Agent" {
		t.Errorf("DisplayName = %q, want 'My Agent'", loaded.CustomAgents[0].DisplayName)
	}
	if loaded.CustomAgents[0].KeyPath != "mcpServers" {
		t.Errorf("KeyPath = %q, want 'mcpServers'", loaded.CustomAgents[0].KeyPath)
	}
}

func TestSaveMCPSyncConfig_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a", "b", "config.json")

	cfg := mcpsyncConfig{Version: 1}
	if err := saveMCPSyncConfig(p, cfg); err != nil {
		t.Fatalf("saveMCPSyncConfig nested dirs: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Error("file should have been created in nested dirs")
	}
}

func TestSaveMCPSyncConfig_OutputIsValidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	cfg := mcpsyncConfig{
		Version: 1,
		CustomAgents: []customAgentEntry{
			{ID: "x", DisplayName: "X", ConfigPath: "/x.json", KeyPath: "mcpServers"},
		},
	}
	if err := saveMCPSyncConfig(p, cfg); err != nil {
		t.Fatalf("saveMCPSyncConfig: %v", err)
	}

	data, _ := os.ReadFile(p)
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("saved config is not valid JSON: %v", err)
	}
}

func TestSaveMCPSyncConfig_TrailingNewline(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	if err := saveMCPSyncConfig(p, mcpsyncConfig{Version: 1}); err != nil {
		t.Fatalf("saveMCPSyncConfig: %v", err)
	}

	data, _ := os.ReadFile(p)
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Error("saved config should end with a newline")
	}
}

// ── resolvePath (cmd package) ─────────────────────────────────────────────────

func TestResolvePath_HomeExpansion(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := resolvePath("~/foo/bar")
	want := filepath.Join(home, "foo", "bar")
	if got != want {
		t.Errorf("resolvePath(~/foo/bar) = %q, want %q", got, want)
	}
}

func TestResolvePath_AbsoluteUnchanged(t *testing.T) {
	abs := "/absolute/path/to/file.json"
	if got := resolvePath(abs); got != abs {
		t.Errorf("resolvePath(absolute) = %q, want %q", got, abs)
	}
}

func TestResolvePath_RelativeUnchanged(t *testing.T) {
	rel := "relative/path"
	if got := resolvePath(rel); got != rel {
		t.Errorf("resolvePath(relative) = %q, want %q", got, rel)
	}
}

func TestResolvePath_TildeNoSlash(t *testing.T) {
	// "~username" should NOT be expanded — only "~/" prefix
	p := "~nodash"
	if got := resolvePath(p); got != p {
		t.Errorf("resolvePath(%q) = %q, want unchanged", p, got)
	}
}

// ── saveMCPSyncConfig file permissions ────────────────────────────────────────

func TestSaveMCPSyncConfig_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	if err := saveMCPSyncConfig(p, mcpsyncConfig{Version: 1}); err != nil {
		t.Fatalf("saveMCPSyncConfig: %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("config file permissions = %o, want 0600", perm)
	}
}
