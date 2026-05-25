package sync

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"github.com/anush-data-portfolio/MCPSync/internal/pathutil"
)

// LoadConfigFile reads a JSON file containing an "mcpServers" object and
// returns the normalized servers it declares. The file shape matches what
// `mcpsync now --path` writes, so configs round-trip cleanly.
//
// The path may use "~" for the home directory.
func LoadConfigFile(path string) ([]agent.NormalizedServer, error) {
	if path == "" {
		return nil, fmt.Errorf("config file path is empty")
	}
	resolved := pathutil.ResolvePath(path)

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", resolved, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", resolved, err)
	}

	serversRaw, ok := raw["mcpServers"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(`%s: missing or invalid "mcpServers" object`, resolved)
	}

	return agent.NormalizeServersMap(serversRaw, "config-file"), nil
}
