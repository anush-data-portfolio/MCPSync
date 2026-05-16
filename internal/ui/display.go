package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"github.com/fatih/color"
)

var jsonKeyRe = regexp.MustCompile(`^(\s*)("[^"]*")(:\s*)(.*)$`)

var (
	green      = color.New(color.FgGreen, color.Bold)
	red        = color.New(color.FgRed, color.Bold)
	yellow     = color.New(color.FgYellow, color.Bold)
	cyan       = color.New(color.FgCyan)
	bold       = color.New(color.Bold)
	dim        = color.New(color.Faint)
	blue       = color.New(color.FgBlue, color.Bold)
	magenta    = color.New(color.FgMagenta)
	serverName = color.New(color.FgCyan, color.Bold)
)

func PrintBanner() {
	bold.Println("\n  MCPSync — MCP Config Sync")
	dim.Println("  Sync MCP servers across all your AI agents")
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
		detail := buildDetail(s)
		fmt.Printf("  %s  %s  %s\n",
			serverName.Sprint(padRight(s.Name, 22)),
			typeColor(s.Type).Sprint(padRight(s.Type, 6)),
			detail,
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
	fmt.Printf("      %s\n", cyan.Sprint("https://github.com/anush-data-portfolio/MCPSync"))
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

func PrintMergedJSON(servers []agent.NormalizedServer, forceJSON bool) {
	if !isTerminal() {
		printMergedRawJSON(servers)
		return
	}
	if forceJSON {
		printMergedColoredJSON(servers)
		return
	}
	printMergedTable(servers)
}

func printMergedColoredJSON(servers []agent.NormalizedServer) {
	mcpServers := buildMCPServersMap(servers)
	out := map[string]any{"mcpServers": mcpServers}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not marshal JSON: %v\n", err)
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		fmt.Println(colorizeJSONKey(line))
	}
}

func colorizeJSONKey(line string) string {
	m := jsonKeyRe.FindStringSubmatch(line)
	if m == nil {
		return line
	}
	indent, key, colon, rest := m[1], m[2], m[3], m[4]
	depth := len(indent) / 2

	var kc *color.Color
	switch depth {
	case 1: // "mcpServers"
		kc = bold
	case 2: // server names like "kaggle-mcp"
		kc = serverName
	case 3: // field keys: "type", "command", "args", "url", "headers"
		kc = yellow
	default: // nested keys: "Authorization", env var names, etc.
		kc = magenta
	}

	return indent + kc.Sprint(key) + colon + rest
}

func isTerminal() bool {
	return !color.NoColor
}

func printMergedTable(servers []agent.NormalizedServer) {
	fmt.Println()
	bold.Printf("  MCP Servers (%d)\n", len(servers))
	fmt.Println()

	for _, s := range servers {
		detail := buildDetail(s)
		extras := buildExtras(s)
		fmt.Printf("  %s  %s  %s%s\n",
			serverName.Sprint(padRight(s.Name, 22)),
			typeColor(s.Type).Sprint(padRight(s.Type, 6)),
			detail,
			extras,
		)
	}
	fmt.Println()
}

func buildDetail(s agent.NormalizedServer) string {
	switch s.Type {
	case "stdio":
		if len(s.Args) > 0 {
			return yellow.Sprint(s.Command) + " " + dim.Sprint(strings.Join(s.Args, " "))
		}
		return yellow.Sprint(s.Command)
	case "http", "sse":
		u := s.URL
		if len(u) > 55 {
			u = u[:52] + "..."
		}
		return cyan.Sprint(u)
	}
	return ""
}

func buildExtras(s agent.NormalizedServer) string {
	var parts []string
	if len(s.Headers) > 0 {
		parts = append(parts, magenta.Sprint("[headers]"))
	}
	if len(s.Env) > 0 {
		parts = append(parts, magenta.Sprint("[env]"))
	}
	if len(parts) == 0 {
		return ""
	}
	return "  " + strings.Join(parts, " ")
}

func typeColor(t string) *color.Color {
	switch t {
	case "stdio":
		return green
	case "http", "sse":
		return blue
	default:
		return dim
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func buildMCPServersMap(servers []agent.NormalizedServer) map[string]any {
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
	return mcpServers
}

func printMergedRawJSON(servers []agent.NormalizedServer) {
	out := map[string]any{"mcpServers": buildMCPServersMap(servers)}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

func WriteMergedJSONToPath(path string, servers []agent.NormalizedServer) error {
	out := map[string]any{"mcpServers": buildMCPServersMap(servers)}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
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
