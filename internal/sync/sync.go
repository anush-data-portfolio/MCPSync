package sync

import (
	"github.com/anush-data-portfolio/MCPSync/internal/agent"
)

func DiscoverAgents(customPaths map[string]string) []agent.AgentInfo {
	infos := make([]agent.AgentInfo, 0, len(agent.AllAgents))

	for _, a := range agent.AllAgents {
		override := customPaths[a.ID()]
		configPath := a.ResolvedConfigPath(override)
		detected := a.IsPresent(configPath)

		serverCount := 0
		if detected {
			result, err := a.Read(configPath)
			if err == nil && result != nil {
				serverCount = len(result.Servers)
			}
		}

		infos = append(infos, agent.AgentInfo{
			ID:          a.ID(),
			DisplayName: a.DisplayName(),
			ConfigPath:  configPath,
			Detected:    detected,
			ServerCount: serverCount,
		})
	}

	return infos
}

func ReadAgents(selectedIDs []string, customPaths map[string]string) ([]agent.AgentReadResult, []error) {
	idSet := make(map[string]bool, len(selectedIDs))
	for _, id := range selectedIDs {
		idSet[id] = true
	}

	var results []agent.AgentReadResult
	var errs []error

	for _, a := range agent.AllAgents {
		if !idSet[a.ID()] {
			continue
		}
		override := customPaths[a.ID()]
		configPath := a.ResolvedConfigPath(override)

		result, err := a.Read(configPath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if result != nil {
			results = append(results, *result)
		}
	}

	return results, errs
}
