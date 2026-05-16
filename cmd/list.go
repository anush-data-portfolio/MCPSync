package cmd

import (
	"fmt"
	"mcpsync/internal/sync"
	"mcpsync/internal/ui"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all detected agents and their config paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.PrintBanner()
		infos := sync.DiscoverAgents(CustomPaths)
		ui.PrintAgentTable(infos)

		detected := 0
		for _, info := range infos {
			if info.Detected {
				detected++
			}
		}

		if detected == 0 {
			fmt.Println("  No agent configs found. To register a custom path:")
			fmt.Println("    mcpsync add --path /your/path/mcp.json")
		} else {
			fmt.Printf("  %d/%d agents detected.\n", detected, len(infos))
			fmt.Println()
			fmt.Println("  To register a custom config path:")
			fmt.Println("    mcpsync add --path /your/path/mcp.json [--name \"My Agent\"]")
		}
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
