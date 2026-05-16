package ui

import (
	"encoding/json"
	"fmt"
	"mcpsync/internal/agent"
	"strings"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen, color.Bold)
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
	cyan   = color.New(color.FgCyan)
	bold   = color.New(color.Bold)
	dim    = color.New(color.Faint)
	blue   = color.New(color.FgBlue, color.Bold)
)

func PrintBanner() {
	bold.Println("\n  MCPSync — MCP Config Sync")
	dim.Println("  Sync MCP servers across all your AI agents\n")
}

func PrintAgentTable(infos []agent.AgentInfo) {
	fmt.Println()
	bold.Println("  Detected agents:")
	fmt.Println()

	for _, info := range infos {
		if info.Detected {
			icon := green.Sprint("✓")
			serverLabel := "servers"
			if info.ServerCount == 1 {
				serverLabel = "server"
			}
			fmt.Printf("  %s  %-28s  %-48s  %s\n",
				icon,
				info.DisplayName,
				dim.Sprint(shortenPath(info.ConfigPath)),
				cyan.Sprintf("%d %s", info.ServerCount, serverLabel),
			)
		} else {
			icon := dim.Sprint("✗")
			fmt.Printf("  %s  %-28s  %s\n",
				icon,
				dim.Sprint(info.DisplayName),
				dim.Sprint("[not found]"),
			)
		}
	}
	fmt.Println()
}

func PrintMergePreview(servers []agent.NormalizedServer) {
	fmt.Println()
	bold.Printf("  Merged: %d unique MCP server(s)\n", len(servers))
	fmt.Println()

	for _, s := range servers {
		detail := ""
		switch s.Type {
		case "stdio":
			args := strings.Join(s.Args, " ")
			if args != "" {
				detail = fmt.Sprintf("%s %s", s.Command, args)
			} else {
				detail = s.Command
			}
		case "http", "sse":
			detail = s.URL
		}
		if len(detail) > 60 {
			detail = detail[:57] + "..."
		}
		fmt.Printf("  %-22s  %-8s  %s\n",
			bold.Sprint(s.Name),
			cyan.Sprint(s.Type),
			dim.Sprint(detail),
		)
	}
	fmt.Println()
}

func PrintConflictHeader(n int) {
	fmt.Println()
	yellow.Printf("  ⚠  %d server name conflict(s) detected — you'll resolve each one now.\n", n)
	fmt.Println()
}

func PrintConflictDetail(c agent.ConflictRecord, idx, total int) {
	rule := strings.Repeat("─", 62)
	fmt.Printf("\n  %s\n", dim.Sprint(rule))
	yellow.Printf("  Conflict %d of %d: %q\n", idx, total, c.ServerName)
	fmt.Printf("  %s\n\n", dim.Sprint(rule))

	printServerBlock("A", c.Kept.SourceAgent, "higher priority", c.Kept)
	fmt.Println()
	printServerBlock("B", c.Discarded.SourceAgent, "lower priority", c.Discarded)

	fmt.Printf("\n  %s\n", dim.Sprint(rule))
}

func printServerBlock(label, source, note string, s agent.NormalizedServer) {
	fmt.Printf("  %s  %s  %s\n",
		blue.Sprintf("[%s]", label),
		bold.Sprint(source),
		dim.Sprintf("(%s)", note),
	)

	fmt.Printf("      %-10s %s\n", dim.Sprint("type:"), cyan.Sprint(s.Type))

	if s.Command != "" {
		fmt.Printf("      %-10s %s\n", dim.Sprint("command:"), s.Command)
	}
	if len(s.Args) > 0 {
		fmt.Printf("      %-10s %s\n", dim.Sprint("args:"), strings.Join(s.Args, " "))
	}
	if s.URL != "" {
		url := s.URL
		if len(url) > 55 {
			url = url[:52] + "..."
		}
		fmt.Printf("      %-10s %s\n", dim.Sprint("url:"), url)
	}
	if len(s.Env) > 0 {
		keys := make([]string, 0, len(s.Env))
		for k := range s.Env {
			keys = append(keys, k)
		}
		fmt.Printf("      %-10s %s\n", dim.Sprint("env:"), strings.Join(keys, ", "))
	}
	if len(s.Headers) > 0 {
		fmt.Printf("      %-10s %s\n", dim.Sprint("headers:"), dim.Sprint("[present]"))
	}
	if len(s.RawExtras) > 0 {
		keys := make([]string, 0, len(s.RawExtras))
		for k := range s.RawExtras {
			keys = append(keys, k)
		}
		fmt.Printf("      %-10s %s\n", dim.Sprint("extras:"), dim.Sprint(strings.Join(keys, ", ")))
	}
}

func PrintStarPrompt() {
	rule := strings.Repeat("─", 62)
	fmt.Println()
	fmt.Printf("  %s\n", dim.Sprint(rule))
	fmt.Printf("  %s  Enjoying MCPSync? Star the repo to support the project:\n", yellow.Sprint("⭐"))
	fmt.Printf("      %s\n", cyan.Sprint("https://github.com/anushkrishnav/mcpsync"))
	fmt.Printf("  %s\n\n", dim.Sprint(rule))
}

func PrintWriteResult(displayName, configPath string, err error, maintain bool) {
	if err != nil {
		red.Printf("  ✗  %-28s  %s\n", displayName, err)
		return
	}
	backup := ""
	if maintain {
		backupPath := strings.TrimSuffix(configPath, ".json") + ".old.json"
		backup = dim.Sprintf("  (backed up → %s)", shortenPath(backupPath))
	}
	green.Printf("  ✓  %-28s", displayName)
	fmt.Printf("  %s%s\n", dim.Sprint(shortenPath(configPath)), backup)
}

func PrintMergedJSON(servers []agent.NormalizedServer) {
	mcpServers := map[string]any{}
	for _, s := range servers {
		entry := map[string]any{"type": s.Type}
		if s.Command != "" {
			entry["command"] = s.Command
		}
		if len(s.Args) > 0 {
			entry["args"] = s.Args
		}
		if len(s.Env) > 0 {
			entry["env"] = s.Env
		}
		if s.URL != "" {
			entry["url"] = s.URL
		}
		if len(s.Headers) > 0 {
			entry["headers"] = s.Headers
		}
		mcpServers[s.Name] = entry
	}

	out := map[string]any{"mcpServers": mcpServers}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(b))
}

func WriteMergedJSONToPath(path string, servers []agent.NormalizedServer) error {
	mcpServers := map[string]any{}
	for _, s := range servers {
		entry := map[string]any{"type": s.Type}
		if s.Command != "" {
			entry["command"] = s.Command
		}
		if len(s.Args) > 0 {
			entry["args"] = s.Args
		}
		if len(s.Env) > 0 {
			entry["env"] = s.Env
		}
		if s.URL != "" {
			entry["url"] = s.URL
		}
		if len(s.Headers) > 0 {
			entry["headers"] = s.Headers
		}
		mcpServers[s.Name] = entry
	}
	out := map[string]any{"mcpServers": mcpServers}

	b, _ := json.MarshalIndent(out, "", "  ")
	b = append(b, '\n')
	return writeFile(path, b)
}

func shortenPath(p string) string {
	home := homeDir()
	if strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}
