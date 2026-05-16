package cmd

import (
	"fmt"
	"os"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"github.com/anush-data-portfolio/MCPSync/internal/merge"
	"github.com/anush-data-portfolio/MCPSync/internal/sync"
	"github.com/anush-data-portfolio/MCPSync/internal/ui"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Show merged MCP servers as a colored table  (use 'get raw' for JSON)",
	Long: `Reads all detected agent configs, merges them into a unique server list,
and displays a colored table. When piped, outputs plain JSON instead.

Use 'mcpsync get raw' to show syntax-highlighted JSON in the terminal.
Conflicts and warnings are printed to stderr so piped output stays clean.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintMergedJSON(loadMergedServers(), false)
	},
}

var getRawCmd = &cobra.Command{
	Use:   "raw",
	Short: "Show the merged MCP server list as syntax-highlighted JSON",
	Long: `Same as 'mcpsync get' but always outputs syntax-highlighted JSON.
When piped, falls back to plain JSON.

Conflicts and warnings are printed to stderr so piped output stays clean.`,
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintMergedJSON(loadMergedServers(), true)
	},
}

func loadMergedServers() []agent.NormalizedServer {
	infos := sync.DiscoverAgents(CustomPaths)

	var selectedIDs []string
	for _, info := range infos {
		if info.Detected {
			selectedIDs = append(selectedIDs, info.ID)
		}
	}

	if len(selectedIDs) == 0 {
		fmt.Fprintln(os.Stderr, "No agent configs found. Run 'mcpsync list' to see expected paths.")
		return nil
	}

	results, errs := sync.ReadAgents(selectedIDs, CustomPaths)
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	merged := merge.Merge(results)
	for _, c := range merged.Conflicts {
		fmt.Fprintf(os.Stderr, "Conflict: %q — kept from %s, discarded from %s\n",
			c.ServerName, c.Kept.SourceAgent, c.Discarded.SourceAgent)
	}

	return merged.Servers
}

func init() {
	getCmd.AddCommand(getRawCmd)
	rootCmd.AddCommand(getCmd)
}
