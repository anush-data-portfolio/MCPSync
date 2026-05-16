package agent

func init() {
	Register(&StandardAgent{
		id:          "claude-desktop",
		displayName: "Claude Desktop",
		defaultPath: "__claude_desktop__", // resolved to platform path in DefaultConfigPath()
		serverKey:   "mcpServers",
	})
}
