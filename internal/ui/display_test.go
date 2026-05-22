package ui

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"github.com/fatih/color"
)

// captureOutput redirects os.Stdout AND color.Output so both fmt.Printf and
// fatih/color functions are captured.
func captureOutput(fn func()) string {
	oldStdout := os.Stdout
	oldColorOut := color.Output
	r, w, err := os.Pipe()
	if err != nil {
		return ""
	}
	os.Stdout = w
	color.Output = w
	fn()
	w.Close()
	os.Stdout = oldStdout
	color.Output = oldColorOut
	b, _ := io.ReadAll(r)
	return string(b)
}

func disableColor(t *testing.T) {
	t.Helper()
	prev := color.NoColor
	color.NoColor = true
	t.Cleanup(func() { color.NoColor = prev })
}

// ── padRight ──────────────────────────────────────────────────────────────────

func TestPadRight(t *testing.T) {
	if got := padRight("hi", 5); got != "hi   " {
		t.Errorf("padRight(%q,5) = %q, want %q", "hi", got, "hi   ")
	}
	if got := padRight("toolong", 4); got != "toolong" {
		t.Errorf("padRight: string longer than width should be returned as-is")
	}
	if got := padRight("exact", 5); got != "exact" {
		t.Errorf("padRight: string at exact width should be returned as-is")
	}
}

// ── typeColor ─────────────────────────────────────────────────────────────────

func TestTypeColor_NotNil(t *testing.T) {
	for _, tc := range []string{"stdio", "http", "sse", "unknown", ""} {
		if typeColor(tc) == nil {
			t.Errorf("typeColor(%q) returned nil", tc)
		}
	}
}

func TestTypeColor_Distinct(t *testing.T) {
	if typeColor("stdio") == typeColor("http") {
		t.Error("stdio and http should have different colors")
	}
}

// ── colorizeJSONKey ───────────────────────────────────────────────────────────

func TestColorizeJSONKey_NonKeyLine(t *testing.T) {
	disableColor(t)
	line := `{`
	if got := colorizeJSONKey(line); got != line {
		t.Errorf("colorizeJSONKey(non-key line) = %q, want %q", got, line)
	}
}

func TestColorizeJSONKey_RetainsStructure(t *testing.T) {
	disableColor(t)
	tests := []struct {
		line  string
		depth int
	}{
		{`  "mcpServers": {`, 1},
		{`    "my-tool": {`, 2},
		{`      "type": "stdio",`, 3},
		{`        "KEY": "value"`, 4},
	}
	for _, tt := range tests {
		got := colorizeJSONKey(tt.line)
		// With NoColor set, ANSI codes are suppressed so output equals input.
		if got != tt.line {
			t.Errorf("colorizeJSONKey(depth %d): got %q, want %q", tt.depth, got, tt.line)
		}
	}
}

// ── buildDetail ───────────────────────────────────────────────────────────────

func TestBuildDetail_StdioWithArgs(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{Type: "stdio", Command: "npx", Args: []string{"-y", "mcp-tool"}}
	got := buildDetail(s)
	if !strings.Contains(got, "npx") {
		t.Errorf("buildDetail stdio+args should contain command, got %q", got)
	}
	if !strings.Contains(got, "-y") || !strings.Contains(got, "mcp-tool") {
		t.Errorf("buildDetail stdio+args should contain args, got %q", got)
	}
}

func TestBuildDetail_StdioNoArgs(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{Type: "stdio", Command: "my-cmd"}
	got := buildDetail(s)
	if !strings.Contains(got, "my-cmd") {
		t.Errorf("buildDetail stdio (no args) should contain command, got %q", got)
	}
}

func TestBuildDetail_HTTP(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{Type: "http", URL: "https://example.com/mcp"}
	got := buildDetail(s)
	if !strings.Contains(got, "https://example.com/mcp") {
		t.Errorf("buildDetail http should contain URL, got %q", got)
	}
}

func TestBuildDetail_LongURL(t *testing.T) {
	disableColor(t)
	longURL := "https://example.com/" + strings.Repeat("a", 50)
	s := agent.NormalizedServer{Type: "http", URL: longURL}
	got := buildDetail(s)
	if !strings.Contains(got, "...") {
		t.Errorf("buildDetail http long URL should be truncated with '...', got %q", got)
	}
	if len(got) > 60 {
		t.Errorf("buildDetail http truncated URL too long: %d chars", len(got))
	}
}

func TestBuildDetail_SSE(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{Type: "sse", URL: "https://example.com/sse"}
	got := buildDetail(s)
	if !strings.Contains(got, "https://example.com/sse") {
		t.Errorf("buildDetail sse should contain URL, got %q", got)
	}
}

func TestBuildDetail_Unknown(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{Type: "other"}
	got := buildDetail(s)
	if got != "" {
		t.Errorf("buildDetail unknown type should return '', got %q", got)
	}
}

// ── buildExtras ───────────────────────────────────────────────────────────────

func TestBuildExtras_None(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{}
	if got := buildExtras(s); got != "" {
		t.Errorf("buildExtras (none) = %q, want ''", got)
	}
}

func TestBuildExtras_HeadersOnly(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{Headers: map[string]string{"Auth": "tok"}}
	got := buildExtras(s)
	if !strings.Contains(got, "[headers]") {
		t.Errorf("buildExtras headers only: want '[headers]', got %q", got)
	}
	if strings.Contains(got, "[env]") {
		t.Errorf("buildExtras headers only: should not contain '[env]', got %q", got)
	}
}

func TestBuildExtras_EnvOnly(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{Env: map[string]string{"KEY": "val"}}
	got := buildExtras(s)
	if !strings.Contains(got, "[env]") {
		t.Errorf("buildExtras env only: want '[env]', got %q", got)
	}
	if strings.Contains(got, "[headers]") {
		t.Errorf("buildExtras env only: should not contain '[headers]', got %q", got)
	}
}

func TestBuildExtras_Both(t *testing.T) {
	disableColor(t)
	s := agent.NormalizedServer{
		Headers: map[string]string{"Auth": "tok"},
		Env:     map[string]string{"KEY": "val"},
	}
	got := buildExtras(s)
	if !strings.Contains(got, "[headers]") || !strings.Contains(got, "[env]") {
		t.Errorf("buildExtras both: want both badges, got %q", got)
	}
}

// ── buildMCPServersMap ────────────────────────────────────────────────────────

func TestBuildMCPServersMap_Minimal(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd"},
	}
	m := buildMCPServersMap(servers, false)
	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
	entry, ok := m["tool"].(map[string]any)
	if !ok {
		t.Fatal("entry is not map[string]any")
	}
	if entry["type"] != "stdio" || entry["command"] != "cmd" {
		t.Errorf("entry fields wrong: %v", entry)
	}
	if _, ok := entry["args"]; ok {
		t.Error("empty args should be omitted")
	}
}

func TestBuildMCPServersMap_AllFields(t *testing.T) {
	servers := []agent.NormalizedServer{
		{
			Name:    "full",
			Type:    "stdio",
			Command: "npx",
			Args:    []string{"-y", "tool"},
			Env:     map[string]string{"KEY": "val"},
		},
	}
	m := buildMCPServersMap(servers, false)
	entry, _ := m["full"].(map[string]any)
	if entry["args"] == nil {
		t.Error("args should be present")
	}
	if entry["env"] == nil {
		t.Error("env should be present")
	}
}

func TestBuildMCPServersMap_HTTP(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "api", Type: "http", URL: "https://example.com"},
	}
	m := buildMCPServersMap(servers, false)
	entry, _ := m["api"].(map[string]any)
	if entry["url"] != "https://example.com" {
		t.Errorf("http entry url wrong: %v", entry["url"])
	}
	if _, ok := entry["command"]; ok {
		t.Error("http entry should not have 'command'")
	}
}

func TestBuildMCPServersMap_Empty(t *testing.T) {
	m := buildMCPServersMap(nil, false)
	if len(m) != 0 {
		t.Errorf("empty input should produce empty map, got %d entries", len(m))
	}
}

// ── WriteMergedJSONToPath ─────────────────────────────────────────────────────

func TestWriteMergedJSONToPath_ValidOutput(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "merged.json")

	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd"},
	}
	if err := WriteMergedJSONToPath(p, servers); err != nil {
		t.Fatalf("WriteMergedJSONToPath: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	data, _ := os.ReadFile(p)
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if out["mcpServers"] == nil {
		t.Error("output missing 'mcpServers' key")
	}
}

func TestWriteMergedJSONToPath_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "dir", "merged.json")

	servers := []agent.NormalizedServer{{Name: "t", Type: "stdio", Command: "c"}}
	if err := WriteMergedJSONToPath(p, servers); err != nil {
		t.Fatalf("WriteMergedJSONToPath nested: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Error("file should have been created in nested directory")
	}
}

func TestWriteMergedJSONToPath_EmptyServers(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty.json")

	if err := WriteMergedJSONToPath(p, nil); err != nil {
		t.Fatalf("WriteMergedJSONToPath(nil): %v", err)
	}
	data, _ := os.ReadFile(p)
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
	mcpServers, _ := out["mcpServers"].(map[string]any)
	if len(mcpServers) != 0 {
		t.Errorf("expected empty mcpServers, got %v", mcpServers)
	}
}

// ── isTerminal ────────────────────────────────────────────────────────────────

func TestIsTerminal_NoColor(t *testing.T) {
	prev := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = prev }()

	if isTerminal() {
		t.Error("isTerminal() should return false when color.NoColor is true")
	}
}

func TestIsTerminal_ColorEnabled(t *testing.T) {
	prev := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = prev }()

	if !isTerminal() {
		t.Error("isTerminal() should return true when color.NoColor is false")
	}
}

// ── shortenPath ───────────────────────────────────────────────────────────────

func TestShortenPath_HomePrefix(t *testing.T) {
	home, _ := os.UserHomeDir()
	p := filepath.Join(home, ".config", "claude", "config.json")
	got := shortenPath(p)
	if !strings.HasPrefix(got, "~") {
		t.Errorf("shortenPath(home-based path) = %q, want '~'-prefixed string", got)
	}
	if strings.Contains(got, home) {
		t.Errorf("shortenPath should replace home dir with '~', got %q", got)
	}
}

func TestShortenPath_NonHomePath(t *testing.T) {
	p := "/etc/some/path.json"
	if got := shortenPath(p); got != p {
		t.Errorf("shortenPath(non-home) = %q, want %q", got, p)
	}
}

func TestShortenPath_ExactHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := shortenPath(filepath.Join(home, "file.json"))
	if !strings.HasPrefix(got, "~") {
		t.Errorf("shortenPath(home + file) = %q, should start with '~'", got)
	}
}

// ── homeDir ───────────────────────────────────────────────────────────────────

func TestHomeDir_NonEmpty(t *testing.T) {
	got := homeDir()
	if got == "" {
		t.Error("homeDir() should return a non-empty string")
	}
}

// ── Print functions (smoke tests via stdout capture) ─────────────────────────

func TestPrintBanner_ContainsMCPSync(t *testing.T) {
	disableColor(t)
	out := captureOutput(PrintBanner)
	if !strings.Contains(out, "MCPSync") {
		t.Errorf("PrintBanner output missing 'MCPSync': %q", out)
	}
}

func TestPrintAgentTable_DetectedAndUndetected(t *testing.T) {
	disableColor(t)
	infos := []agent.AgentInfo{
		{ID: "claude-desktop", DisplayName: "Claude Desktop", ConfigPath: "/foo/bar.json", Detected: true, ServerCount: 3},
		{ID: "cursor", DisplayName: "Cursor", ConfigPath: "/baz.json", Detected: false},
	}
	out := captureOutput(func() { PrintAgentTable(infos) })
	if !strings.Contains(out, "Claude Desktop") {
		t.Errorf("PrintAgentTable: missing 'Claude Desktop' in: %q", out)
	}
	if !strings.Contains(out, "Cursor") {
		t.Errorf("PrintAgentTable: missing 'Cursor' in: %q", out)
	}
}

func TestPrintAgentTable_ServerLabelSingular(t *testing.T) {
	disableColor(t)
	infos := []agent.AgentInfo{
		{ID: "a", DisplayName: "Agent A", ConfigPath: "/a.json", Detected: true, ServerCount: 1},
	}
	out := captureOutput(func() { PrintAgentTable(infos) })
	if !strings.Contains(out, "1 server") {
		t.Errorf("PrintAgentTable: single server should say '1 server', got: %q", out)
	}
}

func TestPrintMergePreview_ContainsServerName(t *testing.T) {
	disableColor(t)
	servers := []agent.NormalizedServer{
		{Name: "my-special-tool", Type: "stdio", Command: "npx"},
	}
	out := captureOutput(func() { PrintMergePreview(servers) })
	if !strings.Contains(out, "my-special-tool") {
		t.Errorf("PrintMergePreview: missing server name in: %q", out)
	}
}

func TestPrintConflictHeader_ShowsCount(t *testing.T) {
	disableColor(t)
	out := captureOutput(func() { PrintConflictHeader(3) })
	if !strings.Contains(out, "3") {
		t.Errorf("PrintConflictHeader(3): missing '3' in: %q", out)
	}
}

func TestPrintConflictDetail_ShowsServerName(t *testing.T) {
	disableColor(t)
	c := agent.ConflictRecord{
		ServerName: "conflict-tool",
		Kept:       agent.NormalizedServer{Name: "conflict-tool", Type: "stdio", Command: "cmd-a", SourceAgent: "agent-a"},
		Discarded:  agent.NormalizedServer{Name: "conflict-tool", Type: "stdio", Command: "cmd-b", SourceAgent: "agent-b"},
	}
	out := captureOutput(func() { PrintConflictDetail(c, 1, 2) })
	if !strings.Contains(out, "conflict-tool") {
		t.Errorf("PrintConflictDetail: missing server name in: %q", out)
	}
}

func TestPrintConflictDetail_AllServerFields(t *testing.T) {
	disableColor(t)
	longURL := "https://example.com/" + strings.Repeat("x", 50)
	c := agent.ConflictRecord{
		ServerName: "full-tool",
		Kept: agent.NormalizedServer{
			Name:        "full-tool",
			Type:        "http",
			URL:         longURL,
			Headers:     map[string]string{"Authorization": "Bearer tok"},
			Env:         map[string]string{"KEY": "val"},
			RawExtras:   map[string]any{"extra": "data"},
			SourceAgent: "agent-a",
		},
		Discarded: agent.NormalizedServer{
			Name:        "full-tool",
			Type:        "stdio",
			Command:     "cmd",
			Args:        []string{"--arg1", "--arg2"},
			SourceAgent: "agent-b",
		},
	}
	// Just verify it doesn't panic and produces output with key fields
	out := captureOutput(func() { PrintConflictDetail(c, 1, 1) })
	if !strings.Contains(out, "full-tool") {
		t.Errorf("PrintConflictDetail all fields: missing server name in: %q", out)
	}
	if !strings.Contains(out, "agent-a") {
		t.Errorf("PrintConflictDetail all fields: missing source agent in: %q", out)
	}
}

func TestPrintStarPrompt_ContainsRepoURL(t *testing.T) {
	disableColor(t)
	out := captureOutput(PrintStarPrompt)
	if !strings.Contains(out, "MCPSync") {
		t.Errorf("PrintStarPrompt: missing 'MCPSync' in: %q", out)
	}
}

func TestPrintWriteResult_Success(t *testing.T) {
	disableColor(t)
	out := captureOutput(func() { PrintWriteResult("My Agent", "/some/path.json", nil, false) })
	if !strings.Contains(out, "My Agent") {
		t.Errorf("PrintWriteResult(success): missing agent name in: %q", out)
	}
}

func TestPrintWriteResult_Error(t *testing.T) {
	disableColor(t)
	out := captureOutput(func() {
		PrintWriteResult("My Agent", "/some/path.json", os.ErrPermission, false)
	})
	if !strings.Contains(out, "My Agent") {
		t.Errorf("PrintWriteResult(error): missing agent name in: %q", out)
	}
}

func TestPrintWriteResult_WithBackup(t *testing.T) {
	disableColor(t)
	out := captureOutput(func() { PrintWriteResult("My Agent", "/path/config.json", nil, true) })
	if !strings.Contains(out, "My Agent") {
		t.Errorf("PrintWriteResult(maintain): missing agent name in: %q", out)
	}
	if !strings.Contains(out, ".old.json") {
		t.Errorf("PrintWriteResult(maintain): missing backup path in: %q", out)
	}
}

// ── PrintMergedJSON routing ───────────────────────────────────────────────────

func TestPrintMergedJSON_NonTerminal_OutputsRawJSON(t *testing.T) {
	disableColor(t) // color.NoColor=true → isTerminal()=false → raw JSON path
	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd"},
	}
	out := captureOutput(func() { PrintMergedJSON(servers, false) })
	if !strings.Contains(out, "mcpServers") {
		t.Errorf("non-terminal PrintMergedJSON should output raw JSON, got: %q", out)
	}
}

func TestPrintMergedJSON_ForceJSON_OutputsJSON(t *testing.T) {
	// With color enabled, forceJSON=true routes to printMergedColoredJSON
	prev := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = prev }()

	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd"},
	}
	out := captureOutput(func() { PrintMergedJSON(servers, true) })
	if !strings.Contains(out, "mcpServers") {
		t.Errorf("forceJSON PrintMergedJSON should contain 'mcpServers', got: %q", out)
	}
}

func TestPrintMergedJSON_Table_ContainsServerName(t *testing.T) {
	// With color enabled and forceJSON=false, routes to printMergedTable
	prev := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = prev }()

	servers := []agent.NormalizedServer{
		{Name: "unique-server-xyz", Type: "stdio", Command: "cmd"},
	}
	out := captureOutput(func() { PrintMergedJSON(servers, false) })
	if !strings.Contains(out, "unique-server-xyz") {
		t.Errorf("table PrintMergedJSON should contain server name, got: %q", out)
	}
}

// ── buildMCPServersMap redaction ──────────────────────────────────────────────

func TestBuildMCPServersMap_Redact_EnvAndHeaders(t *testing.T) {
	servers := []agent.NormalizedServer{
		{
			Name:    "secret",
			Type:    "http",
			URL:     "https://api.example.com",
			Env:     map[string]string{"API_KEY": "super-secret-value"},
			Headers: map[string]string{"Authorization": "Bearer token123"},
		},
	}
	m := buildMCPServersMap(servers, true)
	entry, _ := m["secret"].(map[string]any)

	env, _ := entry["env"].(map[string]string)
	if env["API_KEY"] != "***" {
		t.Errorf("redact=true: env value should be '***', got %q", env["API_KEY"])
	}
	headers, _ := entry["headers"].(map[string]string)
	if headers["Authorization"] != "***" {
		t.Errorf("redact=true: header value should be '***', got %q", headers["Authorization"])
	}
}

func TestBuildMCPServersMap_NoRedact_PreservesValues(t *testing.T) {
	servers := []agent.NormalizedServer{
		{
			Name:    "plain",
			Type:    "stdio",
			Command: "cmd",
			Env:     map[string]string{"MY_VAR": "real-value"},
		},
	}
	m := buildMCPServersMap(servers, false)
	entry, _ := m["plain"].(map[string]any)
	env, _ := entry["env"].(map[string]string)
	if env["MY_VAR"] != "real-value" {
		t.Errorf("redact=false: env value should be preserved, got %q", env["MY_VAR"])
	}
}

// ── WriteMergedJSONToPath file permissions ────────────────────────────────────

func TestWriteMergedJSONToPath_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "merged.json")

	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd"},
	}
	if err := WriteMergedJSONToPath(p, servers); err != nil {
		t.Fatalf("WriteMergedJSONToPath: %v", err)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}
