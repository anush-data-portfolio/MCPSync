package agent

func init() {
	Register(&StandardAgent{
		id:          "gemini",
		displayName: "Gemini CLI",
		defaultPath: "~/.gemini/settings.json",
		serverKey:   "mcpServers",
	})
}
