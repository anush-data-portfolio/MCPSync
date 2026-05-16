package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/anush-data-portfolio/MCPSync/internal/ui"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type customAgentEntry struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	ConfigPath  string `json:"configPath"`
	KeyPath     string `json:"keyPath"`
}

type mcpsyncConfig struct {
	CustomAgents []customAgentEntry `json:"customAgents"`
	Version      int                `json:"version"`
}

var (
	addPath string
	addID   string
	addName string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Register a custom agent config file",
	Long: `Register a custom MCP config file that MCPSync doesn't auto-detect.
The file must be a JSON file containing MCP server definitions.

Example:
  mcpsync add --path /path/to/my-agent/mcp.json --name "My Custom Agent"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resolved := resolvePath(addPath)
		if _, err := os.Stat(resolved); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", resolved)
		}

		data, err := os.ReadFile(resolved)
		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		keyPath := autoDetectKey(raw)
		if keyPath == "" {
			var askErr error
			keyPath, askErr = ui.SelectOne(
				"Could not auto-detect MCP key. Which key contains your servers?",
				[]string{"mcpServers", "servers", "context_servers", "other"},
			)
			if askErr != nil {
				return askErr
			}
			if keyPath == "other" {
				keyPath, askErr = ui.AskInput("Enter the JSON key name:", "mcpServers")
				if askErr != nil {
					return askErr
				}
			}
		}

		displayName := addName
		if displayName == "" {
			base := filepath.Base(filepath.Dir(resolved))
			var askErr error
			displayName, askErr = ui.AskInput("Display name for this agent:", base)
			if askErr != nil {
				return askErr
			}
		}

		agentID := addID
		if agentID == "" {
			agentID = strings.ToLower(strings.ReplaceAll(displayName, " ", "-"))
		}

		cfgPath := mcpsyncConfigPath()
		cfg := loadMCPSyncConfig(cfgPath)

		filtered := cfg.CustomAgents[:0]
		for _, e := range cfg.CustomAgents {
			if e.ID != agentID {
				filtered = append(filtered, e)
			}
		}
		cfg.CustomAgents = filtered
		cfg.CustomAgents = append(cfg.CustomAgents, customAgentEntry{
			ID:          agentID,
			DisplayName: displayName,
			ConfigPath:  resolved,
			KeyPath:     keyPath,
		})
		cfg.Version = 1

		if err := saveMCPSyncConfig(cfgPath, cfg); err != nil {
			return fmt.Errorf("could not save config: %w", err)
		}

		color.Green("\n  ✓  Registered %q (%s)\n", displayName, resolved)
		fmt.Println("  Run 'mcpsync list' to verify, then 'mcpsync now' to sync.\n")
		return nil
	},
}

func autoDetectKey(raw map[string]any) string {
	for _, key := range []string{"mcpServers", "servers", "context_servers"} {
		if _, ok := raw[key]; ok {
			return key
		}
	}
	return ""
}

func mcpsyncConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mcpsync", "config.json")
}

func loadMCPSyncConfig(path string) mcpsyncConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return mcpsyncConfig{}
	}
	var cfg mcpsyncConfig
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveMCPSyncConfig(path string, cfg mcpsyncConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func resolvePath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		p = filepath.Join(home, p[2:])
	}
	return p
}

func init() {
	addCmd.Flags().StringVar(&addPath, "path", "", "Path to the agent MCP config file (required)")
	addCmd.Flags().StringVar(&addID, "id", "", "Agent ID (optional, auto-generated from name)")
	addCmd.Flags().StringVar(&addName, "name", "", "Display name for this agent")
	addCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(addCmd)
}
