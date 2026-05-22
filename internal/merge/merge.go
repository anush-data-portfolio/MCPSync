package merge

import (
	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"sort"
)

// lower index = higher priority in conflict resolution
var agentPriority = []string{
	"claude-desktop", "claude-code", "cursor", "vscode",
	"windsurf", "gemini", "copilot-cli", "copilot-vscode",
	"continue", "junie", "zed",
}

func priorityOf(id string) int {
	for i, p := range agentPriority {
		if p == id {
			return i
		}
	}
	return len(agentPriority) // unknown agents go last
}

func Merge(results []agent.AgentReadResult) agent.MergeResult {
	sorted := make([]agent.AgentReadResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return priorityOf(sorted[i].AgentID) < priorityOf(sorted[j].AgentID)
	})

	seen := map[string]agent.NormalizedServer{}
	var conflicts []agent.ConflictRecord

	for _, result := range sorted {
		for _, server := range result.Servers {
			existing, exists := seen[server.Name]
			if !exists {
				seen[server.Name] = server
				continue
			}
			if !serversEqual(existing, server) {
				conflicts = append(conflicts, agent.ConflictRecord{
					ServerName: server.Name,
					Kept:       existing,
					Discarded:  server,
				})
			}
		}
	}

	servers := make([]agent.NormalizedServer, 0, len(seen))
	for _, s := range seen {
		servers = append(servers, s)
	}
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})

	return agent.MergeResult{Servers: servers, Conflicts: conflicts}
}

func serversEqual(a, b agent.NormalizedServer) bool {
	if a.Type != b.Type || a.Command != b.Command || a.URL != b.URL {
		return false
	}
	if len(a.Args) != len(b.Args) {
		return false
	}
	for i := range a.Args {
		if a.Args[i] != b.Args[i] {
			return false
		}
	}
	if len(a.Env) != len(b.Env) {
		return false
	}
	for k, v := range a.Env {
		if b.Env[k] != v {
			return false
		}
	}
	if len(a.Headers) != len(b.Headers) {
		return false
	}
	for k, v := range a.Headers {
		if b.Headers[k] != v {
			return false
		}
	}
	return true
}
