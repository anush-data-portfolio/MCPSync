package agent

func init() {
	Register(&StandardAgent{
		id:          "cursor",
		displayName: "Cursor",
		defaultPath: "~/.cursor/mcp.json",
		serverKey:   "mcpServers",
	})
}
