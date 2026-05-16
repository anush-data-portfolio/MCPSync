package agent

import "fmt"

type StandardAgent struct {
	id          string
	displayName string
	defaultPath string
	serverKey   string
}

func (a *StandardAgent) ID() string          { return a.id }
func (a *StandardAgent) DisplayName() string { return a.displayName }
func (a *StandardAgent) DefaultConfigPath() string {
	if a.defaultPath == "__claude_desktop__" {
		return claudeDesktopPath()
	}
	return resolvePath(a.defaultPath)
}

func (a *StandardAgent) ResolvedConfigPath(override string) string {
	if override != "" {
		return resolvePath(override)
	}
	return a.DefaultConfigPath()
}

func (a *StandardAgent) IsPresent(configPath string) bool {
	return fileExists(configPath)
}

func (a *StandardAgent) Read(configPath string) (*AgentReadResult, error) {
	raw, err := readJSON(configPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	rawServers, _ := raw[a.serverKey].(map[string]any)
	servers := normalizeServersMap(rawServers, a.id)

	return &AgentReadResult{
		AgentID:   a.id,
		Servers:   servers,
		RawConfig: raw,
	}, nil
}

func (a *StandardAgent) Write(servers []NormalizedServer, configPath string, rawConfig map[string]any, maintain bool) error {
	if rawConfig == nil {
		rawConfig = map[string]any{}
	}
	clone := deepCopyMap(rawConfig)
	clone[a.serverKey] = denormalizeServersSlice(servers, "")

	if err := writeJSON(configPath, clone, maintain); err != nil {
		return fmt.Errorf("%s: %w", a.displayName, err)
	}
	return nil
}
