package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeCodeAgent_Identity(t *testing.T) {
	a := &claudeCodeAgent{}
	if a.ID() != "claude-code" {
		t.Errorf("ID() = %q, want 'claude-code'", a.ID())
	}
	if a.DisplayName() != "Claude Code" {
		t.Errorf("DisplayName() = %q, want 'Claude Code'", a.DisplayName())
	}
}

func TestClaudeCodeAgent_DefaultConfigPath(t *testing.T) {
	a := &claudeCodeAgent{}
	p := a.DefaultConfigPath()
	if p == "" {
		t.Error("DefaultConfigPath() should not be empty")
	}
}

func TestClaudeCodeAgent_ResolvedConfigPath_NoOverride(t *testing.T) {
	a := &claudeCodeAgent{}
	got := a.ResolvedConfigPath("")
	want := a.DefaultConfigPath()
	if got != want {
		t.Errorf("ResolvedConfigPath(\"\") = %q, want %q", got, want)
	}
}

func TestClaudeCodeAgent_ResolvedConfigPath_WithOverride(t *testing.T) {
	a := &claudeCodeAgent{}
	override := "/custom/.claude.json"
	got := a.ResolvedConfigPath(override)
	if got != override {
		t.Errorf("ResolvedConfigPath(override) = %q, want %q", got, override)
	}
}

func TestClaudeCodeAgent_ReadNonExistent(t *testing.T) {
	a := &claudeCodeAgent{}
	result, err := a.Read("/nonexistent/.claude.json")
	if err != nil {
		t.Errorf("Read(nonexistent): want nil error, got %v", err)
	}
	if result != nil {
		t.Error("Read(nonexistent): want nil result")
	}
}

func TestClaudeCodeAgent_ReadDeduplicates(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".claude.json")

	config := map[string]any{
		"projects": map[string]any{
			"/proj1": map[string]any{
				"mcpServers": map[string]any{
					"playwright": map[string]any{"type": "stdio", "command": "npx"},
					"sqlite":     map[string]any{"type": "stdio", "command": "npx"},
				},
			},
			"/proj2": map[string]any{
				"mcpServers": map[string]any{
					"playwright": map[string]any{"type": "stdio", "command": "npx"}, // duplicate
					"extra":      map[string]any{"type": "stdio", "command": "extra"},
				},
			},
		},
	}
	b, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(p, b, 0o644)

	a := &claudeCodeAgent{}
	result, err := a.Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// playwright appears in both projects but should be deduplicated
	if len(result.Servers) != 3 {
		t.Errorf("expected 3 unique servers, got %d", len(result.Servers))
	}
}

func TestClaudeCodeAgent_ReadEmptyProjects(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".claude.json")
	b, _ := json.Marshal(map[string]any{"projects": map[string]any{}})
	os.WriteFile(p, b, 0o644)

	a := &claudeCodeAgent{}
	result, err := a.Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(result.Servers) != 0 {
		t.Errorf("expected 0 servers from empty projects, got %d", len(result.Servers))
	}
}

func TestClaudeCodeAgent_WriteUpdatesAllProjects(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".claude.json")

	config := map[string]any{
		"projects": map[string]any{
			"/proj1": map[string]any{"mcpServers": map[string]any{}},
			"/proj2": map[string]any{"mcpServers": map[string]any{}},
			"/proj3": map[string]any{"mcpServers": map[string]any{}},
		},
	}
	b, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(p, b, 0o644)

	rawConfig, _ := readJSON(p)
	a := &claudeCodeAgent{}
	servers := []NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd"},
	}
	if err := a.Write(servers, p, rawConfig, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Verify all three projects got the server
	raw, _ := readJSON(p)
	projects, _ := raw["projects"].(map[string]any)
	for projPath, projVal := range projects {
		proj, _ := projVal.(map[string]any)
		mcpServers, _ := proj["mcpServers"].(map[string]any)
		if mcpServers["tool"] == nil {
			t.Errorf("project %q did not get 'tool' server", projPath)
		}
	}
}

func TestClaudeCodeAgent_WriteNilRawConfigErrors(t *testing.T) {
	a := &claudeCodeAgent{}
	err := a.Write(nil, "/any/path", nil, false)
	if err == nil {
		t.Error("Write with nil rawConfig should return error (can't write without existing file)")
	}
}

func TestClaudeCodeAgent_IsPresent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".claude.json")
	a := &claudeCodeAgent{}

	if a.IsPresent(p) {
		t.Error("IsPresent should be false before file exists")
	}
	os.WriteFile(p, []byte("{}"), 0o644)
	if !a.IsPresent(p) {
		t.Error("IsPresent should be true after file exists")
	}
}
