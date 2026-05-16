package agent

import "fmt"

type copilotCLIAgent struct{}

func (a *copilotCLIAgent) ID() string          { return "copilot-cli" }
func (a *copilotCLIAgent) DisplayName() string { return "GitHub Copilot CLI" }
func (a *copilotCLIAgent) DefaultConfigPath() string {
	return resolvePath("~/.copilot/mcp-config.json")
}
func (a *copilotCLIAgent) ResolvedConfigPath(override string) string {
	if override != "" {
		return resolvePath(override)
	}
	return a.DefaultConfigPath()
}
func (a *copilotCLIAgent) IsPresent(configPath string) bool { return fileExists(configPath) }

func (a *copilotCLIAgent) Read(configPath string) (*AgentReadResult, error) {
	raw, err := readJSON(configPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	rawServers, _ := raw["mcpServers"].(map[string]any)
	servers := normalizeServersMap(rawServers, a.ID())

	return &AgentReadResult{
		AgentID:   a.ID(),
		Servers:   servers,
		RawConfig: raw,
	}, nil
}

func (a *copilotCLIAgent) Write(servers []NormalizedServer, configPath string, rawConfig map[string]any, maintain bool) error {
	if rawConfig == nil {
		rawConfig = map[string]any{}
	}
	clone := deepCopyMap(rawConfig)

	out := map[string]any{}
	for _, s := range servers {
		typeStr := s.Type
		if typeStr == "stdio" {
			typeStr = "local"
		}
		entry := denormalizeServer(s, typeStr)
		// Copilot CLI schema requires "args" to always be present for local servers.
		if typeStr == "local" {
			if _, hasArgs := entry["args"]; !hasArgs {
				entry["args"] = []string{}
			}
		}
		out[s.Name] = entry
	}
	clone["mcpServers"] = out

	if err := writeJSON(configPath, clone, maintain); err != nil {
		return fmt.Errorf("GitHub Copilot CLI: %w", err)
	}
	return nil
}

func init() {
	Register(&copilotCLIAgent{})
}
