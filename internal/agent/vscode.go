package agent

func init() {
	Register(&StandardAgent{
		id:          "vscode",
		displayName: "VS Code",
		defaultPath: "~/.vscode/mcp.json",
		serverKey:   "mcpServers",
	})
}
