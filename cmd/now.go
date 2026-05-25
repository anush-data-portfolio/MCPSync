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
var useConfigPath string

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "Auto-discover, merge, and sync MCP servers across all agents",
	Long: `Discovers all agent configs, merges their MCP servers into a unique list,
then writes the merged list back to every selected agent.

When two agents have the same server name with different configs, MCPSync
pauses and asks you how to resolve it — keep one, use the other, rename,
exclude, or delete.

Use --use-config <file> to skip discovery/merge and push a single source-of-truth
config (JSON with an "mcpServers" object) to every detected agent. Useful for
agent sandboxes where you want all agents to share an identical tool surface.

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

		var selectedIDs []string
		var mergedServers []agent.NormalizedServer
		var idToResult map[string]agent.AgentReadResult

		if useConfigPath != "" {
			loaded, loadErr := sync.LoadConfigFile(useConfigPath)
			if loadErr != nil {
				return loadErr
			}
			if len(loaded) == 0 {
				fmt.Printf("\n  Config file has no MCP servers — nothing to sync: %s\n", useConfigPath)
				return nil
			}
			sort.Slice(loaded, func(i, j int) bool { return loaded[i].Name < loaded[j].Name })
			mergedServers = loaded

			for _, info := range infos {
				if info.Detected {
					selectedIDs = append(selectedIDs, info.ID)
				}
			}

			// Still read existing configs so non-MCP fields (e.g. other settings in
			// Zed/VSCode) are preserved when writing back.
			results, errs := sync.ReadAgents(selectedIDs, CustomPaths)
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  Warning: %v\n", e)
			}
			idToResult = make(map[string]agent.AgentReadResult, len(results))
			for _, r := range results {
				idToResult[r.AgentID] = r
			}

			fmt.Printf("\n  Using config file: %s\n", useConfigPath)
			fmt.Printf("  Targeting %d detected agent(s).\n", len(selectedIDs))
		} else {
			ids, err := ui.SelectAgents(infos)
			if err != nil {
				return err
			}
			if len(ids) == 0 {
				fmt.Println("\n  No agents selected. Nothing to do.")
				return nil
			}
			selectedIDs = ids

			results, errs := sync.ReadAgents(selectedIDs, CustomPaths)
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "  Warning: %v\n", e)
			}

			merged := merge.Merge(results)

			if len(merged.Conflicts) > 0 {
				ui.PrintConflictHeader(len(merged.Conflicts))
				var resolutions []ui.ConflictResolution
				for i, conflict := range merged.Conflicts {
					res, resolveErr := ui.ResolveConflict(conflict, i+1, len(merged.Conflicts))
					if resolveErr != nil {
						return resolveErr
					}
					resolutions = append(resolutions, res)
				}
				merged.Servers = applyResolutions(merged.Servers, resolutions)
			}

			if len(merged.Servers) == 0 {
				fmt.Println("\n  No MCP servers remaining after conflict resolution.")
				return nil
			}
			mergedServers = merged.Servers

			idToResult = make(map[string]agent.AgentReadResult, len(results))
			for _, r := range results {
				idToResult[r.AgentID] = r
			}
		}

		ui.PrintMergePreview(mergedServers)

		confirmed, err := ui.Confirm(fmt.Sprintf("Write %d server(s) to %d agent(s)?",
			len(mergedServers), len(selectedIDs)))
		if err != nil || !confirmed {
			fmt.Println("\n  Canceled.")
			return nil
		}
		fmt.Println()

		idToInfo := make(map[string]agent.AgentInfo, len(infos))
		for _, info := range infos {
			idToInfo[info.ID] = info
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
			writeErr := a.Write(mergedServers, info.ConfigPath, rawConfig, Maintain)
			ui.PrintWriteResult(a.DisplayName(), info.ConfigPath, writeErr, Maintain)
			if writeErr != nil {
				failed++
			} else {
				written++
			}
		}

		if outputPath != "" {
			if err := ui.WriteMergedJSONToPath(outputPath, mergedServers); err != nil {
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
	nowCmd.Flags().StringVar(&useConfigPath, "use-config", "",
		"Push the MCP servers from this JSON file to every detected agent (skips discovery/merge)")
	rootCmd.AddCommand(nowCmd)
}
