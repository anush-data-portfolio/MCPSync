package ui

import (
	"strings"
	"testing"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
)

// ── summarizeServer ───────────────────────────────────────────────────────────

func TestSummarizeServer_Stdio_WithArgs(t *testing.T) {
	s := agent.NormalizedServer{Type: "stdio", Command: "npx", Args: []string{"-y", "mcp-tool"}}
	got := summarizeServer(s)
	if !strings.Contains(got, "stdio") {
		t.Errorf("summarizeServer stdio+args: missing 'stdio', got %q", got)
	}
	if !strings.Contains(got, "npx") {
		t.Errorf("summarizeServer stdio+args: missing command 'npx', got %q", got)
	}
	if !strings.Contains(got, "-y") || !strings.Contains(got, "mcp-tool") {
		t.Errorf("summarizeServer stdio+args: missing args, got %q", got)
	}
}

func TestSummarizeServer_Stdio_NoArgs(t *testing.T) {
	s := agent.NormalizedServer{Type: "stdio", Command: "my-binary"}
	got := summarizeServer(s)
	if !strings.Contains(got, "stdio") {
		t.Errorf("summarizeServer stdio (no args): missing 'stdio', got %q", got)
	}
	if !strings.Contains(got, "my-binary") {
		t.Errorf("summarizeServer stdio (no args): missing command, got %q", got)
	}
}

func TestSummarizeServer_HTTP(t *testing.T) {
	s := agent.NormalizedServer{Type: "http", URL: "https://example.com/mcp"}
	got := summarizeServer(s)
	if !strings.Contains(got, "http") {
		t.Errorf("summarizeServer http: missing 'http', got %q", got)
	}
	if !strings.Contains(got, "https://example.com/mcp") {
		t.Errorf("summarizeServer http: missing URL, got %q", got)
	}
}

func TestSummarizeServer_SSE(t *testing.T) {
	s := agent.NormalizedServer{Type: "sse", URL: "https://example.com/sse"}
	got := summarizeServer(s)
	if !strings.Contains(got, "sse") {
		t.Errorf("summarizeServer sse: missing 'sse', got %q", got)
	}
}

func TestSummarizeServer_HTTP_LongURL(t *testing.T) {
	longURL := "https://example.com/" + strings.Repeat("x", 50)
	s := agent.NormalizedServer{Type: "http", URL: longURL}
	got := summarizeServer(s)
	if !strings.Contains(got, "...") {
		t.Errorf("summarizeServer long URL should be truncated with '...', got %q", got)
	}
}

func TestSummarizeServer_UnknownType(t *testing.T) {
	s := agent.NormalizedServer{Type: "grpc"}
	got := summarizeServer(s)
	if got != "grpc" {
		t.Errorf("summarizeServer unknown type: got %q, want 'grpc'", got)
	}
}
