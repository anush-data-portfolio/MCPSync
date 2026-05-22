package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func resolvePath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	return filepath.Clean(p)
}

func claudeDesktopPath() string {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "Claude", "claude_desktop_config.json")
	default:
		return resolvePath("~/.config/claude/claude_desktop_config.json")
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func readJSON(p string) (map[string]any, error) {
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", p, err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return out, nil
}

// readJSONC reads a JSONC file (JSON with // and /* */ comments).
// Comments are stripped before parsing; the returned map contains only data.
// Use this for config files known to allow comments (e.g. Zed settings.json).
func readJSONC(p string) (map[string]any, error) {
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", p, err)
	}
	stripped := stripJSONComments(data)
	var out map[string]any
	if err := json.Unmarshal(stripped, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return out, nil
}

// stripJSONComments removes // line comments and /* block comments */ from
// JSON bytes without touching content inside string literals.
func stripJSONComments(src []byte) []byte {
	buf := make([]byte, 0, len(src))
	i := 0
	for i < len(src) {
		c := src[i]

		// Inside a string: copy verbatim, handle escape sequences.
		if c == '"' {
			buf = append(buf, c)
			i++
			for i < len(src) {
				c = src[i]
				buf = append(buf, c)
				i++
				if c == '\\' && i < len(src) {
					// Escaped character — copy the next byte as-is.
					buf = append(buf, src[i])
					i++
				} else if c == '"' {
					break
				}
			}
			continue
		}

		// Single-line comment: skip to end of line.
		if c == '/' && i+1 < len(src) && src[i+1] == '/' {
			i += 2
			for i < len(src) && src[i] != '\n' {
				i++
			}
			continue
		}

		// Block comment: skip to closing */.
		if c == '/' && i+1 < len(src) && src[i+1] == '*' {
			i += 2
			for i+1 < len(src) {
				if src[i] == '*' && src[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		}

		buf = append(buf, c)
		i++
	}
	return bytes.TrimSpace(buf)
}

// writeJSON uses an atomic write (tmp → rename) to prevent corruption.
// It preserves the original file's permissions; new files are created with 0o600.
func writeJSON(p string, data map[string]any, maintain bool) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(p), err)
	}

	perm := os.FileMode(0o600)
	if info, err := os.Stat(p); err == nil {
		perm = info.Mode().Perm()
	}

	if maintain && fileExists(p) {
		backupPath := strings.TrimSuffix(p, ".json") + ".old.json"
		if err := copyFile(p, backupPath, perm); err != nil {
			return fmt.Errorf("backup %s: %w", p, err)
		}
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	content = append(content, '\n')

	tmpPath := p + ".tmp"
	if err := os.WriteFile(tmpPath, content, perm); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, p); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, p, err)
	}
	return nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, perm)
}

func deepCopyMap(m map[string]any) (map[string]any, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("deepCopyMap marshal: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("deepCopyMap unmarshal: %w", err)
	}
	return out, nil
}

func normalizeServer(name string, raw map[string]any, sourceAgent string) NormalizedServer {
	s := NormalizedServer{
		Name:        name,
		SourceAgent: sourceAgent,
	}

	if t, _ := raw["type"].(string); t != "" {
		if t == "local" {
			s.Type = "stdio"
		} else {
			s.Type = t
		}
	}

	s.Command, _ = raw["command"].(string)
	s.URL, _ = raw["url"].(string)

	if args, ok := raw["args"].([]any); ok {
		for _, a := range args {
			if str, ok := a.(string); ok {
				s.Args = append(s.Args, str)
			}
		}
	}

	if env, ok := raw["env"].(map[string]any); ok && len(env) > 0 {
		s.Env = make(map[string]string)
		for k, v := range env {
			if str, ok := v.(string); ok {
				s.Env[k] = str
			}
		}
	}

	if headers, ok := raw["headers"].(map[string]any); ok && len(headers) > 0 {
		s.Headers = make(map[string]string)
		for k, v := range headers {
			if str, ok := v.(string); ok {
				s.Headers[k] = str
			}
		}
	}

	extras := map[string]any{}
	knownKeys := map[string]bool{
		"type": true, "command": true, "args": true,
		"env": true, "url": true, "headers": true,
	}
	for k, v := range raw {
		if !knownKeys[k] {
			extras[k] = v
		}
	}
	if len(extras) > 0 {
		s.RawExtras = extras
	}

	return s
}

func denormalizeServer(s NormalizedServer, typeOverride string) map[string]any {
	out := map[string]any{}

	t := s.Type
	if typeOverride != "" {
		t = typeOverride
	}
	if t != "" {
		out["type"] = t
	}

	if s.Command != "" {
		out["command"] = s.Command
	}
	if len(s.Args) > 0 {
		out["args"] = s.Args
	}
	if len(s.Env) > 0 {
		out["env"] = s.Env
	}
	if s.URL != "" {
		out["url"] = s.URL
	}
	if len(s.Headers) > 0 {
		out["headers"] = s.Headers
	}

	for k, v := range s.RawExtras {
		out[k] = v
	}

	return out
}

func normalizeServersMap(raw map[string]any, sourceAgent string) []NormalizedServer {
	var servers []NormalizedServer
	for name, v := range raw {
		if srv, ok := v.(map[string]any); ok {
			servers = append(servers, normalizeServer(name, srv, sourceAgent))
		}
	}
	return servers
}

func denormalizeServersSlice(servers []NormalizedServer, typeOverride string) map[string]any {
	out := map[string]any{}
	for _, s := range servers {
		out[s.Name] = denormalizeServer(s, typeOverride)
	}
	return out
}
