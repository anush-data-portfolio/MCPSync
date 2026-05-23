package agent

import "fmt"

type antigravityAgent struct{}

func (a *antigravityAgent) ID() string          { return "antigravity" }
func (a *antigravityAgent) DisplayName() string { return "Antigravity (agy)" }
func (a *antigravityAgent) DefaultConfigPath() string {
	return resolvePath("~/.gemini/antigravity/mcp_config.json")
}
func (a *antigravityAgent) ResolvedConfigPath(override string) string {
	if override != "" {
		return resolvePath(override)
	}
	return a.DefaultConfigPath()
}
func (a *antigravityAgent) IsPresent(configPath string) bool { return fileExists(configPath) }

func (a *antigravityAgent) Read(configPath string) (*AgentReadResult, error) {
	raw, err := readJSON(configPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	rawServers, _ := raw["mcpServers"].(map[string]any)
	var servers []NormalizedServer
	for name, v := range rawServers {
		srv, ok := v.(map[string]any)
		if !ok {
			continue
		}
		// Antigravity uses "serverUrl" instead of "url" for HTTP servers.
		if serverURL, ok := srv["serverUrl"].(string); ok && serverURL != "" {
			rewritten := make(map[string]any, len(srv))
			for k, val := range srv {
				rewritten[k] = val
			}
			rewritten["url"] = serverURL
			delete(rewritten, "serverUrl")
			srv = rewritten
		}
		servers = append(servers, normalizeServer(name, srv, a.ID()))
	}

	return &AgentReadResult{
		AgentID:   a.ID(),
		Servers:   servers,
		RawConfig: raw,
	}, nil
}

func (a *antigravityAgent) Write(servers []NormalizedServer, configPath string, rawConfig map[string]any, maintain bool) error {
	if rawConfig == nil {
		rawConfig = map[string]any{}
	}
	clone, err := deepCopyMap(rawConfig)
	if err != nil {
		return fmt.Errorf("Antigravity: copy config: %w", err)
	}

	out := map[string]any{}
	for _, s := range servers {
		entry := denormalizeServer(s, "")
		// Antigravity uses "serverUrl" instead of "url" for HTTP servers.
		if urlVal, ok := entry["url"]; ok {
			entry["serverUrl"] = urlVal
			delete(entry, "url")
		}
		out[s.Name] = entry
	}
	clone["mcpServers"] = out

	if err := writeJSON(configPath, clone, maintain); err != nil {
		return fmt.Errorf("Antigravity: %w", err)
	}
	return nil
}

func init() {
	Register(&antigravityAgent{})
}
