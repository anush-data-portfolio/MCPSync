package agent

func init() {
	Register(&StandardAgent{
		id:          "copilot-vscode",
		displayName: "GitHub Copilot (VS Code)",
		defaultPath: "~/.config/github-copilot/mcp.json",
		serverKey:   "servers", // uses "servers", not "mcpServers"
	})
}
