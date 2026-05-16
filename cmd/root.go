package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Maintain bool
var noMaintain bool
var CustomPaths = map[string]string{}

// version is set at build time via -ldflags "-X github.com/anush-data-portfolio/MCPSync/cmd.version=vX.Y.Z"
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "mcpsync",
	Short: "Sync MCP server configs across AI coding agents",
	Long: `MCPSync collects MCP server configurations from all your AI coding agents,
merges them into a single deduplicated list, and writes back to every agent.

Run 'mcpsync now' to sync everything in one shot.`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&Maintain, "maintain", true,
		"Back up original config files as .old.json before writing")
	// pflag doesn't auto-negate booleans, so --no-maintain is registered explicitly
	rootCmd.PersistentFlags().BoolVar(&noMaintain, "no-maintain", false,
		"Skip .old.json backups before writing")
}

func init() {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if noMaintain {
			Maintain = false
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
