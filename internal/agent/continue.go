package agent

func init() {
	Register(&StandardAgent{
		id:          "continue",
		displayName: "Continue",
		defaultPath: "~/.continue/config.json",
		serverKey:   "mcpServers",
	})
}
