package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZedAgent_Identity(t *testing.T) {
	a := &zedAgent{}
	if a.ID() != "zed" {
		t.Errorf("ID() = %q, want 'zed'", a.ID())
	}
	if a.DisplayName() != "Zed" {
		t.Errorf("DisplayName() = %q, want 'Zed'", a.DisplayName())
	}
}

func TestZedAgent_DefaultConfigPath(t *testing.T) {
	a := &zedAgent{}
	p := a.DefaultConfigPath()
	if p == "" {
		t.Error("DefaultConfigPath() should not be empty")
	}
	if p == "~/.config/zed/settings.json" {
		t.Error("DefaultConfigPath() should return expanded path, not raw '~' path")
	}
}

func TestZedAgent_ResolvedConfigPath_NoOverride(t *testing.T) {
	a := &zedAgent{}
	got := a.ResolvedConfigPath("")
	want := a.DefaultConfigPath()
	if got != want {
		t.Errorf("ResolvedConfigPath(\"\") = %q, want %q", got, want)
	}
}

func TestZedAgent_ResolvedConfigPath_WithOverride(t *testing.T) {
	a := &zedAgent{}
	override := "/custom/zed/settings.json"
	got := a.ResolvedConfigPath(override)
	if got != override {
		t.Errorf("ResolvedConfigPath(override) = %q, want %q", got, override)
	}
}

func TestZedAgent_IsPresent_NonExistent(t *testing.T) {
	a := &zedAgent{}
	if a.IsPresent("/nonexistent/settings.json") {
		t.Error("IsPresent should be false for non-existent file")
	}
}

func TestZedAgent_IsPresent_Existing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	os.WriteFile(p, []byte(`{}`), 0o644)
	a := &zedAgent{}
	if !a.IsPresent(p) {
		t.Error("IsPresent should be true for existing file")
	}
}

func TestZedAgent_ReadJSONC(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	content := `{
  // Zed settings with comments
  "context_servers": {
    /* tool definition */
    "my-tool": {
      "command": {
        "path": "my-binary",
        "args": ["--flag", "--verbose"],
        "env": {"API_KEY": "secret"}
      }
    },
    "bare-tool": {
      "command": {"path": "bare-cmd"}
    }
  },
  "theme": "dark"
}`
	os.WriteFile(p, []byte(content), 0o644)

	a := &zedAgent{}
	result, err := a.Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(result.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(result.Servers))
	}
	if result.AgentID != "zed" {
		t.Errorf("AgentID = %q, want 'zed'", result.AgentID)
	}

	var myTool *NormalizedServer
	for i := range result.Servers {
		if result.Servers[i].Name == "my-tool" {
			myTool = &result.Servers[i]
		}
	}
	if myTool == nil {
		t.Fatal("my-tool not found")
	}
	if myTool.Type != "stdio" {
		t.Errorf("Type = %q, want 'stdio'", myTool.Type)
	}
	if myTool.Command != "my-binary" {
		t.Errorf("Command = %q, want 'my-binary'", myTool.Command)
	}
	if len(myTool.Args) != 2 || myTool.Args[0] != "--flag" {
		t.Errorf("Args = %v, want [--flag --verbose]", myTool.Args)
	}
	if myTool.Env["API_KEY"] != "secret" {
		t.Errorf("Env[API_KEY] = %q, want 'secret'", myTool.Env["API_KEY"])
	}
}

func TestZedAgent_ReadNonExistent(t *testing.T) {
	a := &zedAgent{}
	result, err := a.Read("/no/such/settings.json")
	if err != nil {
		t.Errorf("Read(nonexistent): want nil error, got %v", err)
	}
	if result != nil {
		t.Error("Read(nonexistent): want nil result")
	}
}

func TestZedAgent_WriteStdioOnly(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")

	a := &zedAgent{}
	servers := []NormalizedServer{
		{Name: "stdio-tool", Type: "stdio", Command: "my-cmd", Args: []string{"--arg"}},
		{Name: "http-tool", Type: "http", URL: "https://example.com"},   // skipped
		{Name: "sse-tool", Type: "sse", URL: "https://example.com/sse"}, // skipped
	}
	if err := a.Write(servers, p, nil, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	result, err := a.Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(result.Servers) != 1 {
		t.Errorf("expected 1 server (http/sse skipped), got %d", len(result.Servers))
	}
	if result.Servers[0].Name != "stdio-tool" {
		t.Errorf("expected 'stdio-tool', got %q", result.Servers[0].Name)
	}
}

func TestZedAgent_WriteNestedCommandFormat(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")

	a := &zedAgent{}
	servers := []NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "the-cmd", Args: []string{"-y", "pkg"}, Env: map[string]string{"K": "v"}},
	}
	a.Write(servers, p, nil, false)

	// Read raw to verify Zed nested format
	raw, _ := readJSON(p)
	cs, _ := raw["context_servers"].(map[string]any)
	toolEntry, _ := cs["tool"].(map[string]any)
	cmd, _ := toolEntry["command"].(map[string]any)

	if cmd["path"] != "the-cmd" {
		t.Errorf("command.path = %v, want 'the-cmd'", cmd["path"])
	}
	args, _ := cmd["args"].([]any)
	if len(args) != 2 {
		t.Errorf("command.args length = %d, want 2", len(args))
	}
	env, _ := cmd["env"].(map[string]any)
	if env["K"] != "v" {
		t.Errorf("command.env[K] = %v, want 'v'", env["K"])
	}
}

func TestZedAgent_WritePreservesOtherSettings(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	os.WriteFile(p, []byte(`{"theme":"dark","font_size":14,"context_servers":{}}`), 0o644)

	rawConfig, _ := readJSON(p)
	a := &zedAgent{}
	a.Write([]NormalizedServer{{Name: "tool", Type: "stdio", Command: "cmd"}}, p, rawConfig, false)

	raw, _ := readJSON(p)
	if raw["theme"] != "dark" {
		t.Error("Write should preserve 'theme' setting")
	}
	if raw["font_size"] != float64(14) {
		t.Error("Write should preserve 'font_size' setting")
	}
}

func TestZedAgent_ReadNoContextServers(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")
	os.WriteFile(p, []byte(`{"theme": "light"}`), 0o644)

	a := &zedAgent{}
	result, err := a.Read(p)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	// Should return empty servers, not an error
	if len(result.Servers) != 0 {
		t.Errorf("expected 0 servers when no context_servers, got %d", len(result.Servers))
	}
}
