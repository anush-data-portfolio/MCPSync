package cmd

import (
	"fmt"
	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"github.com/anush-data-portfolio/MCPSync/internal/merge"
	"github.com/anush-data-portfolio/MCPSync/internal/sync"
	"github.com/anush-data-portfolio/MCPSync/internal/ui"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

var outputPath string

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "Auto-discover, merge, and sync MCP servers across all agents",
	Long: `Discovers all agent configs, merges their MCP servers into a unique list,
then writes the merged list back to every selected agent.

When two agents have the same server name with different configs, MCPSync
pauses and asks you how to resolve it — keep one, use the other, rename,
exclude, or delete.

By default, original files are backed up as .old.json before writing.
Use --no-maintain to skip backups.`,
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
			fmt.Println("  No agent configs found. Run 'mcpsync list' for more info.")
			return nil
		}

		selectedIDs, err := ui.SelectAgents(infos)
		if err != nil {
			return err
		}
		if len(selectedIDs) == 0 {
			fmt.Println("\n  No agents selected. Nothing to do.")
			return nil
		}

		results, errs := sync.ReadAgents(selectedIDs, CustomPaths)
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  Warning: %v\n", e)
		}

		merged := merge.Merge(results)

		if len(merged.Conflicts) > 0 {
			ui.PrintConflictHeader(len(merged.Conflicts))
			var resolutions []ui.ConflictResolution
			for i, conflict := range merged.Conflicts {
				res, err := ui.ResolveConflict(conflict, i+1, len(merged.Conflicts))
				if err != nil {
					return err
				}
				resolutions = append(resolutions, res)
			}
			merged.Servers = applyResolutions(merged.Servers, resolutions)
		}

		if len(merged.Servers) == 0 {
			fmt.Println("\n  No MCP servers remaining after conflict resolution.")
			return nil
		}

		ui.PrintMergePreview(merged.Servers)

		confirmed, err := ui.Confirm(fmt.Sprintf("Write %d server(s) to %d agent(s)?",
			len(merged.Servers), len(selectedIDs)))
		if err != nil || !confirmed {
			fmt.Println("\n  Cancelled.")
			return nil
		}
		fmt.Println()

		idToInfo := make(map[string]agent.AgentInfo, len(infos))
		for _, info := range infos {
			idToInfo[info.ID] = info
		}
		idToResult := make(map[string]agent.AgentReadResult, len(results))
		for _, r := range results {
			idToResult[r.AgentID] = r
		}
		selectedSet := make(map[string]bool, len(selectedIDs))
		for _, id := range selectedIDs {
			selectedSet[id] = true
		}

		written, failed := 0, 0
		for _, a := range agent.AllAgents {
			if !selectedSet[a.ID()] {
				continue
			}
			info, ok := idToInfo[a.ID()]
			if !ok {
				continue
			}
			var rawConfig map[string]any
			if r, ok := idToResult[a.ID()]; ok {
				rawConfig = r.RawConfig
			}
			writeErr := a.Write(merged.Servers, info.ConfigPath, rawConfig, Maintain)
			ui.PrintWriteResult(a.DisplayName(), info.ConfigPath, writeErr, Maintain)
			if writeErr != nil {
				failed++
			} else {
				written++
			}
		}

		if outputPath != "" {
			if err := ui.WriteMergedJSONToPath(outputPath, merged.Servers); err != nil {
				fmt.Fprintf(os.Stderr, "\n  Failed to write output file: %v\n", err)
			} else {
				fmt.Printf("\n  Merged config also written to: %s\n", outputPath)
			}
		}

		fmt.Println()
		if failed == 0 {
			fmt.Printf("  Sync complete: %d/%d agents updated.\n", written, written+failed)
		} else {
			fmt.Printf("  Sync complete: %d/%d agents updated (%d failed).\n",
				written, written+failed, failed)
		}
		if Maintain {
			fmt.Println("  Original files backed up as .old.json")
		}

		ui.PrintStarPrompt()
		return nil
	},
}

// "rename" results are appended and re-sorted; "exclude"/"delete" are dropped entirely.
func applyResolutions(servers []agent.NormalizedServer, resolutions []ui.ConflictResolution) []agent.NormalizedServer {
	overwrites := map[string]agent.NormalizedServer{}
	excluded := map[string]bool{}
	var additions []agent.NormalizedServer

	for _, r := range resolutions {
		switch r.Action {
		case "overwrite":
			overwrites[r.ServerName] = r.NewServer
		case "rename":
			additions = append(additions, r.NewServer)
		case "exclude", "delete":
			excluded[r.ServerName] = true
		}
	}

	result := make([]agent.NormalizedServer, 0, len(servers)+len(additions))
	for _, s := range servers {
		if excluded[s.Name] {
			continue
		}
		if repl, ok := overwrites[s.Name]; ok {
			result = append(result, repl)
		} else {
			result = append(result, s)
		}
	}
	result = append(result, additions...)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func init() {
	nowCmd.Flags().StringVarP(&outputPath, "path", "p", "",
		"Also write merged config JSON to this file path")
	rootCmd.AddCommand(nowCmd)
}
