package agent

type NormalizedServer struct {
	Name        string
	Type        string // "stdio", "http", "sse" — never "local"
	Command     string
	Args        []string
	Env         map[string]string
	URL         string
	Headers     map[string]string
	SourceAgent string // not written back to disk
	RawExtras   map[string]any
}

type AgentReadResult struct {
	AgentID   string
	Servers   []NormalizedServer
	RawConfig map[string]any
}

type AgentInfo struct {
	ID          string
	DisplayName string
	ConfigPath  string
	Detected    bool
	ServerCount int
}

type ConflictRecord struct {
	ServerName string
	Kept       NormalizedServer
	Discarded  NormalizedServer
}

type MergeResult struct {
	Servers   []NormalizedServer
	Conflicts []ConflictRecord
}
