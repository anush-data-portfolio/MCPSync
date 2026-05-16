package cmd

import (
	"fmt"
	"mcpsync/internal/merge"
	"mcpsync/internal/sync"
	"mcpsync/internal/ui"
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Print the current merged MCP server list as JSON",
	Long: `Reads all detected agent configs, merges them into a unique server list,
and prints the result as a mcpServers JSON object to stdout.

Conflicts and warnings are printed to stderr so the JSON output can be piped.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Conflicts go to stderr so JSON is still pipeable
		for _, c := range merged.Conflicts {
			fmt.Fprintf(os.Stderr, "Conflict: %q — kept from %s, discarded from %s\n",
				c.ServerName, c.Kept.SourceAgent, c.Discarded.SourceAgent)
		}

		ui.PrintMergedJSON(merged.Servers)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
