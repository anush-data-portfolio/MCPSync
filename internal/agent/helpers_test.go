package agent

import (
	"os"
	"path/filepath"
	"testing"
)

// ── claudeDesktopPath ─────────────────────────────────────────────────────────

func TestClaudeDesktopPath_NonEmpty(t *testing.T) {
	p := claudeDesktopPath()
	if p == "" {
		t.Error("claudeDesktopPath() should return a non-empty path")
	}
}

// ── resolvePath ───────────────────────────────────────────────────────────────

func TestResolvePath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo/bar", filepath.Join(home, "foo", "bar")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~nodash", "~nodash"}, // tilde not followed by slash — not expanded
	}
	for _, tt := range tests {
		got := resolvePath(tt.input)
		if got != tt.want {
			t.Errorf("resolvePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── fileExists ────────────────────────────────────────────────────────────────

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	if fileExists(filepath.Join(dir, "nope.json")) {
		t.Error("fileExists: should be false for missing file")
	}
	p := filepath.Join(dir, "exists.json")
	os.WriteFile(p, []byte("{}"), 0o644)
	if !fileExists(p) {
		t.Error("fileExists: should be true for existing file")
	}
}

// ── stripJSONComments ─────────────────────────────────────────────────────────

func TestStripJSONComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no comments", `{"key": "value"}`, `{"key": "value"}`},
		{"line comment", "{\"k\": 1 // end\n}", "{\"k\": 1 \n}"},
		{"block comment", `{"k": /* note */ 1}`, `{"k":  1}`},
		{"url in string untouched", `{"url": "https://example.com"}`, `{"url": "https://example.com"}`},
		{"double-slash in string untouched", `{"u": "a//b"}`, `{"u": "a//b"}`},
		{"escaped quote in string", `{"k": "say \"hi\""}`, `{"k": "say \"hi\""}`},
		{"multiline block comment", "{\n  /* block\n  comment */\n  \"k\": 1\n}", "{\n  \n  \"k\": 1\n}"},
		{"consecutive line comments", "// first\n// second\n{}", "{}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripJSONComments([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("stripJSONComments:\n  input: %q\n  got:   %q\n  want:  %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── readJSON ──────────────────────────────────────────────────────────────────

func TestReadJSON(t *testing.T) {
	dir := t.TempDir()

	t.Run("nonexistent returns nil nil", func(t *testing.T) {
		got, err := readJSON(filepath.Join(dir, "nope.json"))
		if err != nil || got != nil {
			t.Errorf("readJSON(nonexistent) = %v, %v; want nil, nil", got, err)
		}
	})

	t.Run("valid json", func(t *testing.T) {
		p := filepath.Join(dir, "good.json")
		os.WriteFile(p, []byte(`{"mcpServers":{"tool":{"type":"stdio"}}}`), 0o644)
		got, err := readJSON(p)
		if err != nil {
			t.Fatalf("readJSON error: %v", err)
		}
		if got["mcpServers"] == nil {
			t.Error("expected mcpServers key")
		}
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		p := filepath.Join(dir, "bad.json")
		os.WriteFile(p, []byte(`{not valid`), 0o644)
		_, err := readJSON(p)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

// ── readJSONC ─────────────────────────────────────────────────────────────────

func TestReadJSONC(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "settings.json")

	content := `{
  // Zed settings
  "context_servers": {
    /* block */
    "tool": {"command": {"path": "cmd"}}
  }
}`
	os.WriteFile(p, []byte(content), 0o644)

	got, err := readJSONC(p)
	if err != nil {
		t.Fatalf("readJSONC error: %v", err)
	}
	cs, ok := got["context_servers"].(map[string]any)
	if !ok {
		t.Fatal("readJSONC: context_servers missing")
	}
	if cs["tool"] == nil {
		t.Error("readJSONC: tool entry missing")
	}
}

func TestReadJSONC_Nonexistent(t *testing.T) {
	got, err := readJSONC("/nonexistent/settings.json")
	if err != nil || got != nil {
		t.Errorf("readJSONC(nonexistent) = %v, %v; want nil, nil", got, err)
	}
}

func TestReadJSONC_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.json")
	// Valid JSONC comment syntax but invalid JSON content after stripping
	os.WriteFile(p, []byte(`{ not: valid }`), 0o644)

	_, err := readJSONC(p)
	if err == nil {
		t.Error("readJSONC(invalid JSON): expected error, got nil")
	}
}

// ── writeJSON ─────────────────────────────────────────────────────────────────

func TestWriteJSON_Basic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "out.json")

	data := map[string]any{"mcpServers": map[string]any{}}
	if err := writeJSON(p, data, false); err != nil {
		t.Fatalf("writeJSON: %v", err)
	}
	if !fileExists(p) {
		t.Error("writeJSON should create the file")
	}
	got, _ := readJSON(p)
	if got["mcpServers"] == nil {
		t.Error("written file missing mcpServers")
	}
}

func TestWriteJSON_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a", "b", "c.json")
	if err := writeJSON(p, map[string]any{}, false); err != nil {
		t.Fatalf("writeJSON nested: %v", err)
	}
	if !fileExists(p) {
		t.Error("writeJSON should create parent directories")
	}
}

func TestWriteJSON_MaintainCreatesBackup(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	os.WriteFile(p, []byte(`{"original":true}`), 0o644)

	if err := writeJSON(p, map[string]any{"new": true}, true); err != nil {
		t.Fatalf("writeJSON maintain: %v", err)
	}

	backup := filepath.Join(dir, "config.old.json")
	if !fileExists(backup) {
		t.Error("maintain=true should create .old.json backup")
	}

	backupData, _ := readJSON(backup)
	if backupData["original"] != true {
		t.Error("backup should contain original data")
	}
}

func TestWriteJSON_MaintainFalseNoBackup(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	os.WriteFile(p, []byte(`{"original":true}`), 0o644)

	writeJSON(p, map[string]any{"new": true}, false)

	backup := filepath.Join(dir, "config.old.json")
	if fileExists(backup) {
		t.Error("maintain=false should not create backup")
	}
}

// ── deepCopyMap ───────────────────────────────────────────────────────────────

func TestDeepCopyMap(t *testing.T) {
	original := map[string]any{
		"top": "value",
		"nested": map[string]any{
			"inner": "data",
		},
	}

	cp := deepCopyMap(original)

	// Mutate copy
	cp["top"] = "mutated"
	if original["top"] != "value" {
		t.Error("deepCopyMap: mutating copy affected original top-level")
	}

	// Mutate nested
	if n, ok := cp["nested"].(map[string]any); ok {
		n["inner"] = "changed"
	}
	if n, ok := original["nested"].(map[string]any); ok {
		if n["inner"] == "changed" {
			t.Error("deepCopyMap: mutating copy affected original nested map")
		}
	}
}

// ── normalizeServer ───────────────────────────────────────────────────────────

func TestNormalizeServer(t *testing.T) {
	tests := []struct {
		name        string
		serverName  string
		raw         map[string]any
		sourceAgent string
		checkFn     func(NormalizedServer) error
	}{
		{
			name:       "stdio server",
			serverName: "playwright",
			raw:        map[string]any{"type": "stdio", "command": "npx", "args": []any{"-y", "@playwright/mcp@latest"}},
			checkFn: func(s NormalizedServer) error {
				if s.Type != "stdio" || s.Command != "npx" || len(s.Args) != 2 {
					return &testErr{"stdio fields wrong"}
				}
				return nil
			},
		},
		{
			name:       "local type normalized to stdio",
			serverName: "tool",
			raw:        map[string]any{"type": "local", "command": "cmd"},
			checkFn: func(s NormalizedServer) error {
				if s.Type != "stdio" {
					return &testErr{"'local' should normalize to 'stdio'"}
				}
				return nil
			},
		},
		{
			name:       "http server",
			serverName: "api",
			raw:        map[string]any{"type": "http", "url": "https://example.com/mcp"},
			checkFn: func(s NormalizedServer) error {
				if s.Type != "http" || s.URL != "https://example.com/mcp" {
					return &testErr{"http fields wrong"}
				}
				return nil
			},
		},
		{
			name:       "env parsed",
			serverName: "env-srv",
			raw:        map[string]any{"type": "stdio", "command": "cmd", "env": map[string]any{"KEY": "val"}},
			checkFn: func(s NormalizedServer) error {
				if s.Env["KEY"] != "val" {
					return &testErr{"env not parsed"}
				}
				return nil
			},
		},
		{
			name:       "headers parsed",
			serverName: "hdr-srv",
			raw:        map[string]any{"type": "http", "url": "https://x.com", "headers": map[string]any{"Auth": "Bearer tok"}},
			checkFn: func(s NormalizedServer) error {
				if s.Headers["Auth"] != "Bearer tok" {
					return &testErr{"headers not parsed"}
				}
				return nil
			},
		},
		{
			name:       "extra keys go to RawExtras",
			serverName: "extra",
			raw:        map[string]any{"type": "stdio", "command": "cmd", "tools": "all"},
			checkFn: func(s NormalizedServer) error {
				if s.RawExtras["tools"] != "all" {
					return &testErr{"extra key not in RawExtras"}
				}
				return nil
			},
		},
		{
			name:        "sourceAgent set",
			serverName:  "s",
			raw:         map[string]any{"type": "stdio", "command": "x"},
			sourceAgent: "my-agent",
			checkFn: func(s NormalizedServer) error {
				if s.SourceAgent != "my-agent" {
					return &testErr{"sourceAgent not set"}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeServer(tt.serverName, tt.raw, tt.sourceAgent)
			if got.Name != tt.serverName {
				t.Errorf("Name = %q, want %q", got.Name, tt.serverName)
			}
			if err := tt.checkFn(got); err != nil {
				t.Error(err)
			}
		})
	}
}

// ── denormalizeServer ─────────────────────────────────────────────────────────

func TestDenormalizeServer(t *testing.T) {
	t.Run("stdio fields", func(t *testing.T) {
		s := NormalizedServer{Type: "stdio", Command: "npx", Args: []string{"-y", "tool"}}
		out := denormalizeServer(s, "")
		if out["type"] != "stdio" || out["command"] != "npx" {
			t.Errorf("wrong fields: %v", out)
		}
		args, _ := out["args"].([]string)
		if len(args) != 2 {
			t.Errorf("args length = %d, want 2", len(args))
		}
	})

	t.Run("type override", func(t *testing.T) {
		s := NormalizedServer{Type: "stdio", Command: "tool"}
		out := denormalizeServer(s, "local")
		if out["type"] != "local" {
			t.Errorf("type override: got %q, want 'local'", out["type"])
		}
	})

	t.Run("empty fields omitted", func(t *testing.T) {
		s := NormalizedServer{Type: "stdio", Command: "cmd"}
		out := denormalizeServer(s, "")
		if _, ok := out["args"]; ok {
			t.Error("empty args should be omitted")
		}
		if _, ok := out["env"]; ok {
			t.Error("nil env should be omitted")
		}
		if _, ok := out["url"]; ok {
			t.Error("empty url should be omitted")
		}
	})

	t.Run("raw extras preserved", func(t *testing.T) {
		s := NormalizedServer{Type: "stdio", Command: "cmd", RawExtras: map[string]any{"tools": "all"}}
		out := denormalizeServer(s, "")
		if out["tools"] != "all" {
			t.Errorf("RawExtras not preserved: %v", out)
		}
	})
}

// ── normalizeServersMap ───────────────────────────────────────────────────────

func TestNormalizeServersMap(t *testing.T) {
	raw := map[string]any{
		"tool1": map[string]any{"type": "stdio", "command": "cmd1"},
		"tool2": map[string]any{"type": "http", "url": "https://example.com"},
		"bad":   "not a map", // should be skipped
	}
	servers := normalizeServersMap(raw, "test-agent")
	if len(servers) != 2 {
		t.Errorf("expected 2 servers (skipping bad entry), got %d", len(servers))
	}
	for _, s := range servers {
		if s.SourceAgent != "test-agent" {
			t.Errorf("SourceAgent = %q, want 'test-agent'", s.SourceAgent)
		}
	}
}

// ── denormalizeServersSlice ───────────────────────────────────────────────────

func TestDenormalizeServersSlice(t *testing.T) {
	servers := []NormalizedServer{
		{Name: "tool1", Type: "stdio", Command: "cmd1"},
		{Name: "tool2", Type: "http", URL: "https://example.com"},
	}
	out := denormalizeServersSlice(servers, "")
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	t1, _ := out["tool1"].(map[string]any)
	if t1["command"] != "cmd1" {
		t.Errorf("tool1.command = %v, want 'cmd1'", t1["command"])
	}
	t2, _ := out["tool2"].(map[string]any)
	if t2["url"] != "https://example.com" {
		t.Errorf("tool2.url = %v, want 'https://example.com'", t2["url"])
	}
}

func TestDenormalizeServersSlice_TypeOverride(t *testing.T) {
	servers := []NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd"},
	}
	out := denormalizeServersSlice(servers, "local")
	entry, _ := out["tool"].(map[string]any)
	if entry["type"] != "local" {
		t.Errorf("type override not applied: %v", entry["type"])
	}
}

// helper

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }
