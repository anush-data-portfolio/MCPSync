package agent

type Agent interface {
	ID() string
	DisplayName() string
	DefaultConfigPath() string
	ResolvedConfigPath(override string) string
	IsPresent(configPath string) bool
	Read(configPath string) (*AgentReadResult, error)
	Write(servers []NormalizedServer, configPath string, rawConfig map[string]any, maintain bool) error
}
