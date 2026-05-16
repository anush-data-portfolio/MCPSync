package agent

import "fmt"

// zedAgent uses "context_servers" with a nested command schema; http/sse not supported.
type zedAgent struct{}

func (a *zedAgent) ID() string          { return "zed" }
func (a *zedAgent) DisplayName() string { return "Zed" }
func (a *zedAgent) DefaultConfigPath() string {
	return resolvePath("~/.config/zed/settings.json")
}
func (a *zedAgent) ResolvedConfigPath(override string) string {
	if override != "" {
		return resolvePath(override)
	}
	return a.DefaultConfigPath()
}
func (a *zedAgent) IsPresent(configPath string) bool { return fileExists(configPath) }

func (a *zedAgent) Read(configPath string) (*AgentReadResult, error) {
	raw, err := readJSON(configPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	contextServers, _ := raw["context_servers"].(map[string]any)
	var servers []NormalizedServer

	for name, val := range contextServers {
		entry, ok := val.(map[string]any)
		if !ok {
			continue
		}
		cmd, _ := entry["command"].(map[string]any)
		s := NormalizedServer{
			Name:        name,
			Type:        "stdio",
			SourceAgent: "zed",
		}
		if cmd != nil {
			s.Command, _ = cmd["path"].(string)
			if args, ok := cmd["args"].([]any); ok {
				for _, a := range args {
					if str, ok := a.(string); ok {
						s.Args = append(s.Args, str)
					}
				}
			}
			if env, ok := cmd["env"].(map[string]any); ok && len(env) > 0 {
				s.Env = make(map[string]string)
				for k, v := range env {
					if str, ok := v.(string); ok {
						s.Env[k] = str
					}
				}
			}
		}
		servers = append(servers, s)
	}

	return &AgentReadResult{
		AgentID:   "zed",
		Servers:   servers,
		RawConfig: raw,
	}, nil
}

func (a *zedAgent) Write(servers []NormalizedServer, configPath string, rawConfig map[string]any, maintain bool) error {
	if rawConfig == nil {
		rawConfig = map[string]any{}
	}
	clone := deepCopyMap(rawConfig)

	contextServers := map[string]any{}
	var skipped int

	for _, s := range servers {
		if s.Type == "http" || s.Type == "sse" {
			skipped++
			continue
		}
		// Zed stdio format: { "command": { "path": "...", "args": [], "env": {} } }
		cmd := map[string]any{}
		if s.Command != "" {
			cmd["path"] = s.Command
		}
		if len(s.Args) > 0 {
			cmd["args"] = s.Args
		}
		if len(s.Env) > 0 {
			cmd["env"] = s.Env
		}
		contextServers[s.Name] = map[string]any{"command": cmd}
	}

	if skipped > 0 {
		fmt.Printf("  [Zed] Skipped %d http/sse server(s) — Zed only supports stdio\n", skipped)
	}

	clone["context_servers"] = contextServers

	if err := writeJSON(configPath, clone, maintain); err != nil {
		return fmt.Errorf("Zed: %w", err)
	}
	return nil
}

func init() {
	Register(&zedAgent{})
}
