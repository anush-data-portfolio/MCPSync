package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFile_EmptyPath(t *testing.T) {
	if _, err := LoadConfigFile(""); err == nil {
		t.Error("LoadConfigFile(\"\"): expected error, got nil")
	}
}

func TestLoadConfigFile_MissingFile(t *testing.T) {
	if _, err := LoadConfigFile("/definitely/does/not/exist.json"); err == nil {
		t.Error("LoadConfigFile(missing): expected error, got nil")
	}
}

func TestLoadConfigFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(p, []byte("not json at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfigFile(p); err == nil {
		t.Error("LoadConfigFile(invalid json): expected error, got nil")
	}
}

func TestLoadConfigFile_MissingMCPServersKey(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nokey.json")
	if err := os.WriteFile(p, []byte(`{"otherKey": {}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfigFile(p); err == nil {
		t.Error("LoadConfigFile(no mcpServers): expected error, got nil")
	}
}

func TestLoadConfigFile_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "ok.json")
	cfg := map[string]any{
		"mcpServers": map[string]any{
			"playwright": map[string]any{
				"type":    "stdio",
				"command": "npx",
				"args":    []any{"@playwright/mcp"},
				"env":     map[string]any{"FOO": "bar"},
			},
			"weather": map[string]any{
				"type": "http",
				"url":  "https://example.com/mcp",
			},
		},
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}

	servers, err := LoadConfigFile(p)
	if err != nil {
		t.Fatalf("LoadConfigFile: unexpected error: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	byName := map[string]bool{}
	for _, s := range servers {
		byName[s.Name] = true
		if s.SourceAgent != "config-file" {
			t.Errorf("server %q: SourceAgent = %q, want \"config-file\"", s.Name, s.SourceAgent)
		}
	}
	if !byName["playwright"] || !byName["weather"] {
		t.Errorf("missing expected server names: got %v", byName)
	}
}

func TestLoadConfigFile_EmptyMCPServers(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(p, []byte(`{"mcpServers": {}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	servers, err := LoadConfigFile(p)
	if err != nil {
		t.Fatalf("LoadConfigFile: unexpected error: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("expected 0 servers for empty mcpServers, got %d", len(servers))
	}
}
