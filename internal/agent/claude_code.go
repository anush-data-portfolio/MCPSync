package agent

import (
	"fmt"
	"os"
	"path/filepath"
)

type claudeCodeAgent struct{}

func (a *claudeCodeAgent) ID() string          { return "claude-code" }
func (a *claudeCodeAgent) DisplayName() string { return "Claude Code" }

func (a *claudeCodeAgent) DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude.json")
}

func (a *claudeCodeAgent) ResolvedConfigPath(override string) string {
	if override != "" {
		return resolvePath(override)
	}
	return a.DefaultConfigPath()
}

func (a *claudeCodeAgent) IsPresent(configPath string) bool {
	return fileExists(configPath)
}

func (a *claudeCodeAgent) Read(configPath string) (*AgentReadResult, error) {
	raw, err := readJSON(configPath)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	projects, _ := raw["projects"].(map[string]any)
	seen := map[string]NormalizedServer{}

	for _, projVal := range projects {
		proj, ok := projVal.(map[string]any)
		if !ok {
			continue
		}
		mcpServers, _ := proj["mcpServers"].(map[string]any)
		for name, srvVal := range mcpServers {
			if _, exists := seen[name]; exists {
				continue
			}
			if srv, ok := srvVal.(map[string]any); ok {
				seen[name] = normalizeServer(name, srv, "claude-code")
			}
		}
	}

	servers := make([]NormalizedServer, 0, len(seen))
	for _, s := range seen {
		servers = append(servers, s)
	}

	return &AgentReadResult{
		AgentID:   "claude-code",
		Servers:   servers,
		RawConfig: raw,
	}, nil
}

func (a *claudeCodeAgent) Write(servers []NormalizedServer, configPath string, rawConfig map[string]any, maintain bool) error {
	if rawConfig == nil {
		return fmt.Errorf("Claude Code: cannot write — config file not found at %s", configPath)
	}

	clone, err := deepCopyMap(rawConfig)
	if err != nil {
		return fmt.Errorf("Claude Code: copy config: %w", err)
	}
	projects, _ := clone["projects"].(map[string]any)
	if projects == nil {
		projects = map[string]any{}
		clone["projects"] = projects
	}

	denormalized := denormalizeServersSlice(servers, "")
	for projPath := range projects {
		proj, ok := projects[projPath].(map[string]any)
		if !ok {
			continue
		}
		proj["mcpServers"] = denormalized
	}

	if err := writeJSON(configPath, clone, maintain); err != nil {
		return fmt.Errorf("Claude Code: %w", err)
	}
	return nil
}

func init() {
	Register(&claudeCodeAgent{})
}
