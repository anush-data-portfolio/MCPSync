package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
)

// agentIDs returns the set of IDs from a slice of AgentInfo.
func agentIDs(infos []agent.AgentInfo) map[string]bool {
	ids := make(map[string]bool, len(infos))
	for _, info := range infos {
		ids[info.ID] = true
	}
	return ids
}

// ── DiscoverAgents ────────────────────────────────────────────────────────────

func TestDiscoverAgents_ReturnsAllRegisteredAgents(t *testing.T) {
	infos := DiscoverAgents(nil)
	if len(infos) == 0 {
		t.Fatal("DiscoverAgents should return at least one agent")
	}
	if len(infos) != len(agent.AllAgents) {
		t.Errorf("expected %d infos (one per registered agent), got %d",
			len(agent.AllAgents), len(infos))
	}
}

func TestDiscoverAgents_InfoHasRequiredFields(t *testing.T) {
	infos := DiscoverAgents(nil)
	for _, info := range infos {
		if info.ID == "" {
			t.Error("AgentInfo.ID must not be empty")
		}
		if info.DisplayName == "" {
			t.Errorf("AgentInfo[%q].DisplayName must not be empty", info.ID)
		}
		if info.ConfigPath == "" {
			t.Errorf("AgentInfo[%q].ConfigPath must not be empty", info.ID)
		}
	}
}

func TestDiscoverAgents_NonExistentFiles_NotDetected(t *testing.T) {
	// All agents will have non-existent config files in CI, so Detected=false.
	// We just check that ServerCount is 0 when Detected is false.
	infos := DiscoverAgents(nil)
	for _, info := range infos {
		if !info.Detected && info.ServerCount != 0 {
			t.Errorf("AgentInfo[%q]: Detected=false but ServerCount=%d", info.ID, info.ServerCount)
		}
	}
}

func TestDiscoverAgents_CustomPath_Detected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp.json")

	// Write a minimal MCP config that a StandardAgent can read
	config := map[string]any{
		"mcpServers": map[string]any{
			"tool1": map[string]any{"type": "stdio", "command": "cmd1"},
			"tool2": map[string]any{"type": "stdio", "command": "cmd2"},
		},
	}
	b, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(p, b, 0o644)

	// Use claude-desktop as the target since it's a StandardAgent
	customPaths := map[string]string{"claude-desktop": p}
	infos := DiscoverAgents(customPaths)

	ids := agentIDs(infos)
	if !ids["claude-desktop"] {
		t.Fatal("claude-desktop should appear in infos")
	}

	for _, info := range infos {
		if info.ID == "claude-desktop" {
			if !info.Detected {
				t.Error("claude-desktop should be detected (file exists)")
			}
			if info.ServerCount != 2 {
				t.Errorf("claude-desktop: expected 2 servers, got %d", info.ServerCount)
			}
			break
		}
	}
}

func TestDiscoverAgents_UnknownCustomPath_Ignored(t *testing.T) {
	customPaths := map[string]string{"nonexistent-agent-id": "/some/path.json"}
	infos := DiscoverAgents(customPaths)
	if len(infos) == 0 {
		t.Error("DiscoverAgents should still return known agents even with unknown custom paths")
	}
}

// ── ReadAgents ────────────────────────────────────────────────────────────────

func TestReadAgents_EmptySelectedIDs(t *testing.T) {
	results, errs := ReadAgents(nil, nil)
	if len(results) != 0 {
		t.Errorf("ReadAgents(nil): expected 0 results, got %d", len(results))
	}
	if len(errs) != 0 {
		t.Errorf("ReadAgents(nil): expected 0 errors, got %d", len(errs))
	}
}

func TestReadAgents_UnknownID_NoResults(t *testing.T) {
	results, errs := ReadAgents([]string{"nonexistent-agent"}, nil)
	if len(results) != 0 {
		t.Errorf("ReadAgents(unknown): expected 0 results, got %d", len(results))
	}
	if len(errs) != 0 {
		t.Errorf("ReadAgents(unknown): expected 0 errors, got %d", len(errs))
	}
}

func TestReadAgents_NonExistentConfig_NoError(t *testing.T) {
	// Agents whose config files don't exist return nil (not an error)
	var someID string
	for _, a := range agent.AllAgents {
		someID = a.ID()
		break
	}

	results, errs := ReadAgents([]string{someID}, map[string]string{someID: "/nonexistent/path.json"})
	if len(errs) != 0 {
		t.Errorf("ReadAgents(nonexistent path): expected 0 errors, got %d: %v", len(errs), errs)
	}
	// nil result from Read means the agent is skipped — so 0 results expected
	if len(results) != 0 {
		t.Errorf("ReadAgents(nonexistent): expected 0 results, got %d", len(results))
	}
}

func TestReadAgents_ExistingConfig_ReturnsResult(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp.json")

	config := map[string]any{
		"mcpServers": map[string]any{
			"playwright": map[string]any{"type": "stdio", "command": "npx"},
		},
	}
	b, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(p, b, 0o644)

	customPaths := map[string]string{"claude-desktop": p}
	results, errs := ReadAgents([]string{"claude-desktop"}, customPaths)
	if len(errs) != 0 {
		t.Fatalf("ReadAgents: unexpected errors: %v", errs)
	}
	if len(results) != 1 {
		t.Fatalf("ReadAgents: expected 1 result, got %d", len(results))
	}
	if results[0].AgentID != "claude-desktop" {
		t.Errorf("AgentID = %q, want 'claude-desktop'", results[0].AgentID)
	}
	if len(results[0].Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(results[0].Servers))
	}
}

func TestReadAgents_MultipleAgents(t *testing.T) {
	dir := t.TempDir()
	config := map[string]any{
		"mcpServers": map[string]any{
			"tool": map[string]any{"type": "stdio", "command": "cmd"},
		},
	}
	b, _ := json.MarshalIndent(config, "", "  ")

	// Write configs for two known StandardAgents
	p1 := filepath.Join(dir, "desktop.json")
	p2 := filepath.Join(dir, "cursor.json")
	os.WriteFile(p1, b, 0o644)
	os.WriteFile(p2, b, 0o644)

	customPaths := map[string]string{
		"claude-desktop": p1,
		"cursor":         p2,
	}
	results, errs := ReadAgents([]string{"claude-desktop", "cursor"}, customPaths)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
