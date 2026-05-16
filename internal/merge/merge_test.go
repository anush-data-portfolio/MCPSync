package merge

import (
	"testing"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
)

// ── priorityOf ────────────────────────────────────────────────────────────────

func TestPriorityOf_KnownAgents(t *testing.T) {
	order := []string{
		"claude-desktop", "claude-code", "cursor", "vscode",
		"windsurf", "gemini", "copilot-cli", "copilot-vscode",
		"continue", "junie", "zed",
	}
	for i, id := range order {
		got := priorityOf(id)
		if got != i {
			t.Errorf("priorityOf(%q) = %d, want %d", id, got, i)
		}
	}
}

func TestPriorityOf_UnknownAgentGoesLast(t *testing.T) {
	unknown := priorityOf("totally-unknown-agent")
	last := priorityOf("zed")
	if unknown <= last {
		t.Errorf("unknown agent priority %d should be > last known (%d)", unknown, last)
	}
}

// ── serversEqual ─────────────────────────────────────────────────────────────

func TestServersEqual(t *testing.T) {
	base := agent.NormalizedServer{Type: "stdio", Command: "npx", Args: []string{"-y", "@playwright/mcp@latest"}}

	tests := []struct {
		name string
		a, b agent.NormalizedServer
		want bool
	}{
		{"identical", base, base, true},
		{"different type", base, agent.NormalizedServer{Type: "http", Command: "npx", Args: []string{"-y", "@playwright/mcp@latest"}}, false},
		{"different command", base, agent.NormalizedServer{Type: "stdio", Command: "node", Args: []string{"-y", "@playwright/mcp@latest"}}, false},
		{"different args len", base, agent.NormalizedServer{Type: "stdio", Command: "npx", Args: []string{"-y"}}, false},
		{"different arg value", base, agent.NormalizedServer{Type: "stdio", Command: "npx", Args: []string{"-y", "other"}}, false},
		{"nil vs empty args", agent.NormalizedServer{Type: "stdio", Command: "cmd"}, agent.NormalizedServer{Type: "stdio", Command: "cmd", Args: []string{}}, true},
		{"http equal", agent.NormalizedServer{Type: "http", URL: "https://a.com"}, agent.NormalizedServer{Type: "http", URL: "https://a.com"}, true},
		{"http different url", agent.NormalizedServer{Type: "http", URL: "https://a.com"}, agent.NormalizedServer{Type: "http", URL: "https://b.com"}, false},
		{"sse type", agent.NormalizedServer{Type: "sse", URL: "https://x.com"}, agent.NormalizedServer{Type: "sse", URL: "https://x.com"}, true},
		{"env/headers ignored", base, agent.NormalizedServer{Type: "stdio", Command: "npx", Args: []string{"-y", "@playwright/mcp@latest"}, Env: map[string]string{"K": "v"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serversEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("serversEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ── Merge ─────────────────────────────────────────────────────────────────────

func TestMerge_Empty(t *testing.T) {
	r := Merge(nil)
	if len(r.Servers) != 0 || len(r.Conflicts) != 0 {
		t.Errorf("Merge(nil): want empty result, got %+v", r)
	}
}

func TestMerge_SingleAgent(t *testing.T) {
	input := []agent.AgentReadResult{{
		AgentID: "cursor",
		Servers: []agent.NormalizedServer{
			{Name: "playwright", Type: "stdio", Command: "npx"},
			{Name: "sqlite", Type: "stdio", Command: "npx"},
		},
	}}
	r := Merge(input)
	if len(r.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(r.Servers))
	}
	if len(r.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(r.Conflicts))
	}
}

func TestMerge_DisjointAgents_NoConflict(t *testing.T) {
	input := []agent.AgentReadResult{
		{AgentID: "claude-desktop", Servers: []agent.NormalizedServer{{Name: "playwright", Type: "stdio", Command: "npx"}}},
		{AgentID: "cursor", Servers: []agent.NormalizedServer{{Name: "sqlite", Type: "stdio", Command: "npx"}}},
	}
	r := Merge(input)
	if len(r.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(r.Servers))
	}
	if len(r.Conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(r.Conflicts))
	}
}

func TestMerge_IdenticalServerNoDuplicate(t *testing.T) {
	s := agent.NormalizedServer{Name: "playwright", Type: "stdio", Command: "npx", Args: []string{"-y", "@playwright/mcp@latest"}}
	input := []agent.AgentReadResult{
		{AgentID: "claude-desktop", Servers: []agent.NormalizedServer{s}},
		{AgentID: "cursor", Servers: []agent.NormalizedServer{s}},
		{AgentID: "vscode", Servers: []agent.NormalizedServer{s}},
	}
	r := Merge(input)
	if len(r.Servers) != 1 {
		t.Errorf("identical servers should deduplicate to 1, got %d", len(r.Servers))
	}
	if len(r.Conflicts) != 0 {
		t.Errorf("identical servers should produce no conflicts, got %d", len(r.Conflicts))
	}
}

func TestMerge_ConflictHigherPriorityWins(t *testing.T) {
	// claude-desktop (priority 0) vs cursor (priority 2) — desktop wins
	input := []agent.AgentReadResult{
		{AgentID: "cursor", Servers: []agent.NormalizedServer{{Name: "tool", Type: "stdio", Command: "cursor-cmd"}}},
		{AgentID: "claude-desktop", Servers: []agent.NormalizedServer{{Name: "tool", Type: "stdio", Command: "desktop-cmd"}}},
	}
	r := Merge(input)
	if len(r.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(r.Servers))
	}
	if r.Servers[0].Command != "desktop-cmd" {
		t.Errorf("claude-desktop should win conflict, got command %q", r.Servers[0].Command)
	}
	if len(r.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(r.Conflicts))
	}
	if r.Conflicts[0].Kept.Command != "desktop-cmd" {
		t.Errorf("Conflict.Kept should be desktop-cmd, got %q", r.Conflicts[0].Kept.Command)
	}
	if r.Conflicts[0].Discarded.Command != "cursor-cmd" {
		t.Errorf("Conflict.Discarded should be cursor-cmd, got %q", r.Conflicts[0].Discarded.Command)
	}
}

func TestMerge_MultipleConflicts(t *testing.T) {
	input := []agent.AgentReadResult{
		{AgentID: "claude-desktop", Servers: []agent.NormalizedServer{
			{Name: "tool1", Type: "stdio", Command: "v1"},
			{Name: "tool2", Type: "http", URL: "https://v1.com"},
		}},
		{AgentID: "cursor", Servers: []agent.NormalizedServer{
			{Name: "tool1", Type: "stdio", Command: "v2"},
			{Name: "tool2", Type: "http", URL: "https://v2.com"},
		}},
	}
	r := Merge(input)
	if len(r.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(r.Servers))
	}
	if len(r.Conflicts) != 2 {
		t.Errorf("expected 2 conflicts, got %d", len(r.Conflicts))
	}
}

func TestMerge_OutputSortedAlphabetically(t *testing.T) {
	input := []agent.AgentReadResult{{
		AgentID: "cursor",
		Servers: []agent.NormalizedServer{
			{Name: "zebra", Type: "stdio"},
			{Name: "alpha", Type: "stdio"},
			{Name: "middle", Type: "stdio"},
		},
	}}
	r := Merge(input)
	for i := 1; i < len(r.Servers); i++ {
		if r.Servers[i].Name < r.Servers[i-1].Name {
			t.Errorf("output not sorted at index %d: %q < %q", i, r.Servers[i].Name, r.Servers[i-1].Name)
		}
	}
}

func TestMerge_UnknownAgentLowestPriority(t *testing.T) {
	input := []agent.AgentReadResult{
		{AgentID: "unknown-custom", Servers: []agent.NormalizedServer{{Name: "tool", Type: "stdio", Command: "custom-cmd"}}},
		{AgentID: "claude-desktop", Servers: []agent.NormalizedServer{{Name: "tool", Type: "stdio", Command: "desktop-cmd"}}},
	}
	r := Merge(input)
	if len(r.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(r.Servers))
	}
	if r.Servers[0].Command != "desktop-cmd" {
		t.Errorf("desktop should beat unknown agent, got %q", r.Servers[0].Command)
	}
}
