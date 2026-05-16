package agent

func init() {
	Register(&StandardAgent{
		id:          "junie",
		displayName: "Junie",
		defaultPath: "~/.junie/mcp/mcp.json",
		serverKey:   "mcpServers",
	})
}
