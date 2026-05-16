package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTestConfig(t *testing.T, dir, filename string, data map[string]any) string {
	t.Helper()
	p := filepath.Join(dir, filename)
	b, _ := json.MarshalIndent(data, "", "  ")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatalf("writeTestConfig: %v", err)
	}
	return p
}

func TestStandardAgent_ID_And_DisplayName(t *testing.T) {
	a := &StandardAgent{id: "my-agent", displayName: "My Agent"}
	if a.ID() != "my-agent" {
		t.Errorf("ID() = %q, want 'my-agent'", a.ID())
	}
	if a.DisplayName() != "My Agent" {
		t.Errorf("DisplayName() = %q, want 'My Agent'", a.DisplayName())
	}
}

func TestStandardAgent_ReadWrite_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	p := writeTestConfig(t, dir, "mcp.json", map[string]any{
		"mcpServers": map[string]any{
			"playwright": map[string]any{"type": "stdio", "command": "npx", "args": []any{"-y", "@playwright/mcp@latest"}},
			"sqlite":     map[string]any{"type": "stdio", "command": "npx", "args": []any{"-y", "mcp-sqlite"}},
		},
	})

	a := &StandardAgent{id: "test", displayName: "Test", serverKey: "mcpServers"}

	result, err := a.Read(p)
	if err != nil || result == nil {
		t.Fatalf("Read: err=%v result=%v", err, result)
	}
	if len(result.Servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(result.Servers))
	}

	err = a.Write(result.Servers, p, result.RawConfig, false)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	result2, err := a.Read(p)
	if err != nil || result2 == nil {
		t.Fatalf("Read after Write: err=%v", err)
	}
	if len(result2.Servers) != 2 {
		t.Errorf("after roundtrip: expected 2 servers, got %d", len(result2.Servers))
	}
}

func TestStandardAgent_ReadNonExistent(t *testing.T) {
	a := &StandardAgent{id: "test", serverKey: "mcpServers"}
	result, err := a.Read("/no/such/path.json")
	if err != nil {
		t.Errorf("Read(nonexistent) should return nil error, got %v", err)
	}
	if result != nil {
		t.Errorf("Read(nonexistent) should return nil result")
	}
}

func TestStandardAgent_ReadEmptyServers(t *testing.T) {
	dir := t.TempDir()
	p := writeTestConfig(t, dir, "mcp.json", map[string]any{
		"mcpServers": map[string]any{},
	})
	a := &StandardAgent{id: "test", serverKey: "mcpServers"}
	result, err := a.Read(p)
	if err != nil || result == nil {
		t.Fatalf("Read(empty): err=%v", err)
	}
	if len(result.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(result.Servers))
	}
}

func TestStandardAgent_IsPresent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp.json")
	a := &StandardAgent{}

	if a.IsPresent(p) {
		t.Error("IsPresent: should be false before file exists")
	}
	os.WriteFile(p, []byte("{}"), 0o644)
	if !a.IsPresent(p) {
		t.Error("IsPresent: should be true after file is created")
	}
}

func TestStandardAgent_WriteCreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "new.json")
	a := &StandardAgent{id: "test", displayName: "Test", serverKey: "mcpServers"}

	servers := []NormalizedServer{{Name: "tool", Type: "stdio", Command: "cmd"}}
	if err := a.Write(servers, p, nil, false); err != nil {
		t.Fatalf("Write: %v", err)
	}

	result, err := a.Read(p)
	if err != nil || result == nil || len(result.Servers) != 1 {
		t.Errorf("Write/Read: err=%v servers=%d", err, len(result.Servers))
	}
}

func TestStandardAgent_WritePreservesExtraKeys(t *testing.T) {
	dir := t.TempDir()
	p := writeTestConfig(t, dir, "mcp.json", map[string]any{
		"mcpServers":  map[string]any{},
		"otherConfig": "preserved",
		"nested":      map[string]any{"key": "value"},
	})

	a := &StandardAgent{id: "test", displayName: "Test", serverKey: "mcpServers"}
	result, _ := a.Read(p)
	servers := []NormalizedServer{{Name: "tool", Type: "stdio", Command: "cmd"}}
	a.Write(servers, p, result.RawConfig, false)

	raw, _ := readJSON(p)
	if raw["otherConfig"] != "preserved" {
		t.Error("Write should preserve non-mcpServers keys")
	}
}

func TestStandardAgent_DifferentServerKey(t *testing.T) {
	dir := t.TempDir()
	p := writeTestConfig(t, dir, "config.json", map[string]any{
		"servers": map[string]any{
			"tool": map[string]any{"type": "stdio", "command": "cmd"},
		},
	})
	a := &StandardAgent{id: "test", serverKey: "servers"}
	result, err := a.Read(p)
	if err != nil || result == nil {
		t.Fatalf("Read with 'servers' key: err=%v", err)
	}
	if len(result.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(result.Servers))
	}
}

func TestStandardAgent_ResolvedConfigPath(t *testing.T) {
	a := &StandardAgent{id: "test", defaultPath: "~/.test/mcp.json"}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".test", "mcp.json")

	got := a.ResolvedConfigPath("")
	if got != expected {
		t.Errorf("ResolvedConfigPath() = %q, want %q", got, expected)
	}

	override := "/custom/path.json"
	got2 := a.ResolvedConfigPath(override)
	if got2 != override {
		t.Errorf("ResolvedConfigPath(override) = %q, want %q", got2, override)
	}
}

func TestStandardAgent_MaintainBackup(t *testing.T) {
	dir := t.TempDir()
	p := writeTestConfig(t, dir, "mcp.json", map[string]any{
		"mcpServers": map[string]any{
			"original": map[string]any{"type": "stdio", "command": "orig-cmd"},
		},
	})

	a := &StandardAgent{id: "test", displayName: "Test", serverKey: "mcpServers"}
	result, _ := a.Read(p)
	newServers := []NormalizedServer{{Name: "new", Type: "stdio", Command: "new-cmd"}}
	a.Write(newServers, p, result.RawConfig, true)

	backup := filepath.Join(dir, "mcp.old.json")
	if !fileExists(backup) {
		t.Error("maintain=true should create .old.json backup")
	}
	backupData, _ := readJSON(backup)
	origServers, _ := backupData["mcpServers"].(map[string]any)
	if origServers["original"] == nil {
		t.Error("backup should contain original server")
	}
}
