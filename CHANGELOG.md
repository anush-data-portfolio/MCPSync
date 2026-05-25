# Changelog

All notable changes to MCPSync are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [0.1.6] — 2026-05-25

### Added

- **`mcpsync now --use-config <file>`** — push a single source-of-truth MCP config file to every detected agent in one shot. The file uses the same `{"mcpServers": {...}}` shape that `mcpsync now --path` produces, so configs round-trip cleanly. Discovery merge and conflict-resolution are skipped; every detected agent's existing non-MCP fields are preserved. Intended for agent sandboxes where you want every agent to share an identical tool surface.

---

## [0.1.4] — 2026-05-16

### Fixed

- **`go install` path**: correct install command is now `go install github.com/anush-data-portfolio/MCPSync/mcpsync@latest`; Makefile and README updated accordingly.
- **GitHub Copilot CLI config**: always write `"args": []` for no-argument stdio servers — resolves `Invalid MCP server configuration: args: Required` error on startup.

---

## [0.1.0] — 2026-05-15

### Added

**Core sync engine**
- `mcpsync now` — auto-discover all agents, merge MCP servers, write back in one command
- `mcpsync get` — print merged MCP server list as JSON (stdout-safe for piping)
- `mcpsync list` — display all supported agents with detection status and server counts
- `mcpsync add --path <file>` — register custom agent config files not auto-detected

**Agent support (12 agents)**
- Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json`) — macOS/Windows/Linux
- Claude Code (`~/.claude.json`) — reads from all projects, writes back to all projects
- Cursor (`~/.cursor/mcp.json`)
- VS Code (`~/.vscode/mcp.json`)
- Windsurf (`~/.codeium/windsurf/mcp_config.json`)
- Zed (`~/.config/zed/settings.json`) — `context_servers` format with stdio-only support
- Continue (`~/.continue/config.json`)
- Gemini CLI (`~/.gemini/settings.json`)
- GitHub Copilot CLI (`~/.copilot/mcp-config.json`) — handles `type: "local"` ↔ `"stdio"` normalization
- GitHub Copilot VS Code (`~/.config/github-copilot/mcp.json`) — `servers` key variant
- OpenAI Codex CLI (`~/.codex/config.json`)
- Junie (`~/.junie/mcp/mcp.json`)

**Merge behavior**
- Deduplication by server name — identical configs silently merged
- **Interactive conflict resolution** — when two agents have the same server name with different configs, MCPSync pauses and asks the user to resolve it with five options:
  - **Keep** — use the higher-priority agent's version (default behavior)
  - **Overwrite** — use the challenger agent's version instead
  - **Rename** — keep both; assign a new name to the challenger's version
  - **Exclude** — skip this server entirely (not written to any agent this sync)
  - **Delete** — remove from the merged list (effectively deleted from all synced agents)
- Each conflict displays a side-by-side view of both server configs (type, command, args, url, env, headers, extras) before asking
- Priority ordering still determines which version is pre-selected as "A" (higher priority): Claude Desktop > Claude Code > Cursor > VS Code > Windsurf > Gemini > Copilot CLI > Copilot VS Code > Continue > Junie > Zed
- Alphabetical sort of merged output for deterministic results

**Safety features**
- Automatic backup of original files as `.old.json` before every write (default on)
- `--no-maintain` flag to opt out of backups
- Atomic file writes (write to `.tmp`, then `rename`) — no partial file corruption
- Dry-run preview before confirming write
- Per-agent error reporting — one failure doesn't stop other agents

**UX**
- Interactive multi-select to choose which agents to include
- Colorized table output (detected ✓ vs not found ✗)
- Merged server preview showing name, type, and endpoint/command
- Side-by-side conflict display before interactive resolution
- ⭐ Star-the-repo prompt shown after a successful sync
- `-p / --path` flag on `now` to additionally save merged JSON to a custom path

**Distribution**
- Single binary — no runtime required after build
- **npm package** — `npm install -g mcpsync` or `npx mcpsync`; downloads the correct pre-built binary for the user's platform automatically via `scripts/postinstall.js`
- **Go install** — `go install github.com/anush-data-portfolio/MCPSync@latest`
- **Pre-built binaries** — darwin/linux/windows via GitHub Releases
- Cross-compile script for all platforms in `scripts/build.sh`
- GitHub Actions workflow (`.github/workflows/release.yml`) builds all platform binaries and publishes to npm on every `v*` tag push

---

## Roadmap

- [ ] `mcpsync diff` — show what would change before syncing
- [ ] `mcpsync restore` — restore from `.old.json` backups
- [ ] `mcpsync ignore <server>` — exclude specific servers from sync
- [ ] Per-agent exclusions — skip writing to specific agents
- [ ] Watch mode — auto-sync when any config file changes
- [ ] Config profiles — named sets of agents for different sync scenarios
- [ ] `mcpsync add --agent cursor` — point a named built-in agent to a custom path
