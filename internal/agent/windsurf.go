package agent

func init() {
	Register(&StandardAgent{
		id:          "windsurf",
		displayName: "Windsurf",
		defaultPath: "~/.codeium/windsurf/mcp_config.json",
		serverKey:   "mcpServers",
	})
}
