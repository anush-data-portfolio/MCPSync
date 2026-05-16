package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCopilotCLIAgent_Identity(t *testing.T) {
	a := &copilotCLIAgent{}
	if a.ID() != "copilot-cli" {
		t.Errorf("ID() = %q, want 'copilot-cli'", a.ID())
	}
	if a.DisplayName() != "GitHub Copilot CLI" {
		t.Errorf("DisplayName() = %q, want 'GitHub Copilot CLI'", a.DisplayName())
	}
}

func TestCopilotCLIAgent_DefaultConfigPath(t *testing.T) {
	a := &copilotCLIAgent{}
	p := a.DefaultConfigPath()
	if p == "" {
		t.Error("DefaultConfigPath() should not be empty")
	}
}

func TestCopilotCLIAgent_ResolvedConfigPath_NoOverride(t *testing.T) {
	a := &copilotCLIAgent{}
	got := a.ResolvedConfigPath("")
	want := a.DefaultConfigPath()
	if got != want {
		t.Errorf("ResolvedConfigPath(\"\") = %q, want %q", got, want)
	}
}

func TestCopilotCLIAgent_ResolvedConfigPath_WithOverride(t *testing.T) {
	a := &copilotCLIAgent{}
	override := "/custom/mcp-config.json"
	got := a.ResolvedConfigPath(override)
	if got != override {
		t.Errorf("ResolvedConfigPath(override) = %q, want %q", got, override)
	}
}

func TestCopilotCLIAgent_IsPresent_NonExistent(t *testing.T) {
	a := &copilotCLIAgent{}
	if a.IsPresent("/nonexistent/mcp-config.json") {
		t.Error("IsPresent should be false for non-existent file")
	}
}

func TestCopilotCLIAgent_IsPresent_Existing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp-config.json")
	os.WriteFile(p, []byte(`{}`), 0o644)
	a := &copilotCLIAgent{}
	if !a.IsPresent(p) {
		t.Error("IsPresent should be true for existing file")
	}
}

func TestCopilotCLIAgent_ReadNormalizesLocalToStdio(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp-config.json")

	config := map[string]any{
		"mcpServers": map[string]any{
			"my-tool": map[string]any{
				"type":    "local",
				"command": "tool-cmd",
				"args":    []any{"--flag"},
			},
		},
	}
	b, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(p, b, 0o644)

	a := &copilotCLIAgent{}
	result, err := a.Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(result.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(result.Servers))
	}
	if result.Servers[0].Type != "stdio" {
		t.Errorf("'local' type should normalize to 'stdio', got %q", result.Servers[0].Type)
	}
	if result.Servers[0].Command != "tool-cmd" {
		t.Errorf("Command = %q, want 'tool-cmd'", result.Servers[0].Command)
	}
}

func TestCopilotCLIAgent_WriteConvertsStdioToLocal(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp-config.json")

	a := &copilotCLIAgent{}
	servers := []NormalizedServer{
		{Name: "my-tool", Type: "stdio", Command: "tool-cmd"},
	}
	if err := a.Write(servers, p, nil, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	raw, _ := readJSON(p)
	mcpServers, _ := raw["mcpServers"].(map[string]any)
	myTool, _ := mcpServers["my-tool"].(map[string]any)
	if myTool["type"] != "local" {
		t.Errorf("Write should convert 'stdio' to 'local' for Copilot CLI, got %q", myTool["type"])
	}
}

func TestCopilotCLIAgent_WriteAlwaysIncludesArgs(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp-config.json")

	a := &copilotCLIAgent{}
	// Server with no args — Copilot CLI schema requires "args" to be present.
	servers := []NormalizedServer{
		{Name: "no-args-tool", Type: "stdio", Command: "some-cmd"},
	}
	if err := a.Write(servers, p, nil, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	raw, _ := readJSON(p)
	mcpServers, _ := raw["mcpServers"].(map[string]any)
	tool, _ := mcpServers["no-args-tool"].(map[string]any)
	if _, ok := tool["args"]; !ok {
		t.Error("Write should always include 'args' for local servers (Copilot CLI schema requires it)")
	}
}

func TestCopilotCLIAgent_WriteUntypedCommandServerGetsLocalAndArgs(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp-config.json")

	a := &copilotCLIAgent{}
	// Server with no type but with a command (e.g., synced from Claude Desktop).
	// Must get type="local" and args=[] so Copilot CLI schema is satisfied.
	servers := []NormalizedServer{
		{Name: "untyped-tool", Type: "", Command: "task-manager"},
	}
	if err := a.Write(servers, p, nil, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	raw, _ := readJSON(p)
	mcpServers, _ := raw["mcpServers"].(map[string]any)
	tool, _ := mcpServers["untyped-tool"].(map[string]any)
	if tool["type"] != "local" {
		t.Errorf("untyped command server should get type='local', got %q", tool["type"])
	}
	if _, ok := tool["args"]; !ok {
		t.Error("untyped command server should get 'args' (Copilot CLI schema requires it)")
	}
}

func TestCopilotCLIAgent_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp-config.json")

	a := &copilotCLIAgent{}
	servers := []NormalizedServer{
		{Name: "tool1", Type: "stdio", Command: "cmd1", Args: []string{"--arg1", "--arg2"}},
		{Name: "tool2", Type: "stdio", Command: "cmd2"},
	}
	if err := a.Write(servers, p, nil, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	result, err := a.Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(result.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(result.Servers))
	}
	for _, s := range result.Servers {
		if s.Type != "stdio" {
			t.Errorf("server %q should be normalized to 'stdio', got %q", s.Name, s.Type)
		}
	}
}

func TestCopilotCLIAgent_ReadNonExistent(t *testing.T) {
	a := &copilotCLIAgent{}
	result, err := a.Read("/nonexistent/mcp-config.json")
	if err != nil {
		t.Errorf("Read(nonexistent) should return nil error, got %v", err)
	}
	if result != nil {
		t.Error("Read(nonexistent) should return nil result")
	}
}

func TestCopilotCLIAgent_WritePreservesNonMCPKeys(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp-config.json")

	// Pre-existing config with non-MCP keys
	existing := map[string]any{
		"mcpServers":  map[string]any{},
		"copilotMode": "advanced",
	}
	b, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(p, b, 0o644)

	rawConfig, _ := readJSON(p)
	a := &copilotCLIAgent{}
	servers := []NormalizedServer{{Name: "tool", Type: "stdio", Command: "cmd"}}
	a.Write(servers, p, rawConfig, false)

	raw, _ := readJSON(p)
	if raw["copilotMode"] != "advanced" {
		t.Error("Write should preserve non-mcpServers keys")
	}
}
