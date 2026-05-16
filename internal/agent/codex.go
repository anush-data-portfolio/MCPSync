package agent

func init() {
	Register(&StandardAgent{
		id:          "codex",
		displayName: "OpenAI Codex CLI",
		defaultPath: "~/.codex/config.json",
		serverKey:   "mcpServers",
	})
}
