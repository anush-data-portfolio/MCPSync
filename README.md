# MCPSync

**One command to sync MCP server configs across all your AI coding agents.**

If you use multiple AI agents — Claude Desktop, Claude Code, Cursor, VS Code, Windsurf, Gemini CLI, Copilot, Codex, Zed, and more — you know the pain: you add an MCP server to one agent and forget all the others. MCPSync fixes that.

```bash
mcpsync now
```

MCPSync reads every agent's config, merges the MCP servers into one unique list, and writes back to all of them. One command. Done.

---

> **Open source** — contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).
> Found an agent we don't support? [Open an agent request →](https://github.com/anushkrishnav/mcpsync/issues/new?template=agent-request.yml)

---

## Quick Start

Pick the option that matches your setup — **no runtime required** once installed.

### Option 1 — npm (no Go needed)

```bash
npm install -g mcpsync
mcpsync now
```

Or with npx (no install at all):
```bash
npx mcpsync list
```

> The npm package downloads the correct pre-built binary for your platform automatically.
> Requires Node.js 18+. No Go installation needed.

### Option 2 — Go install

```bash
go install github.com/anushkrishnav/mcpsync@latest
mcpsync now
```

### Option 3 — Download a pre-built binary

Download from [Releases](https://github.com/anushkrishnav/mcpsync/releases), make executable, and move to your PATH:

```bash
# macOS (Apple Silicon)
curl -Lo mcpsync https://github.com/anushkrishnav/mcpsync/releases/latest/download/mcpsync-darwin-arm64
chmod +x mcpsync && sudo mv mcpsync /usr/local/bin/

# macOS (Intel)
curl -Lo mcpsync https://github.com/anushkrishnav/mcpsync/releases/latest/download/mcpsync-darwin-amd64
chmod +x mcpsync && sudo mv mcpsync /usr/local/bin/

# Linux (x86_64)
curl -Lo mcpsync https://github.com/anushkrishnav/mcpsync/releases/latest/download/mcpsync-linux-amd64
chmod +x mcpsync && sudo mv mcpsync /usr/local/bin/
```

### Option 4 — Build from source

```bash
git clone https://github.com/anushkrishnav/mcpsync
cd mcpsync
go build -o mcpsync .
sudo mv mcpsync /usr/local/bin/
mcpsync now
```

**Requirements:** Go 1.21+ (for building only — the binary runs without Go)

---

## Commands

### `mcpsync now`

Auto-discover all agents, merge MCP servers, write back to all.

```bash
mcpsync now                  # Interactive: select agents, preview, confirm, sync
mcpsync now --no-maintain    # Skip .old.json backups
mcpsync now -p output.json   # Also save merged config to a file
```

**Flow:**
1. Detects all agent configs on your machine
2. Shows a multi-select so you can choose which agents to include
3. Merges all selected servers into a single deduplicated list
4. Shows a preview — server names, types, and commands/URLs
5. Asks for confirmation, then writes to every selected agent
6. Backs up every modified file as `filename.old.json` (on by default)

### `mcpsync get`

Print the current merged MCP server list as JSON. Safe to pipe.

```bash
mcpsync get
mcpsync get > merged.json
mcpsync get | pbcopy      # copy to clipboard (macOS)
```

Conflicts and warnings go to stderr, so the JSON on stdout stays clean.

### `mcpsync list`

Show all supported agents and whether their configs were found.

```
  ✓  Claude Desktop          ~/Library/…/claude_desktop_config.json    3 servers
  ✓  Claude Code             ~/.claude.json                             2 servers
  ✓  GitHub Copilot CLI      ~/.copilot/mcp-config.json                2 servers
  ✗  Cursor                  ~/.cursor/mcp.json                        [not found]
  ✗  VS Code                 ~/.vscode/mcp.json                        [not found]
  ...
```

### `mcpsync add --path <file>`

Register a custom agent config that MCPSync didn't auto-detect.

```bash
mcpsync add --path ~/my-custom-agent/mcp.json
mcpsync add --path ~/my-custom-agent/mcp.json --name "My Agent"
```

MCPSync saves it to `~/.mcpsync/config.json` and includes it in future syncs.

---

## Supported Agents

MCPSync currently supports **12 agents** out of the box. If yours isn't listed, [request it](https://github.com/anushkrishnav/mcpsync/issues/new?template=agent-request.yml) or add it yourself — it's usually ~10 lines of Go.

### Claude (Anthropic)

| Variant | Config File | Notes |
|---------|-------------|-------|
| **Claude Desktop** | `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) | The Anthropic desktop app. Windows: `%APPDATA%\Claude\claude_desktop_config.json` |
| **Claude Code** | `~/.claude.json` | The Claude CLI for coding. Stores MCP per-project; MCPSync reads from all projects and writes back to all. |

### AI Coding Agents / IDEs

| Agent | Config File | Key | Notes |
|-------|-------------|-----|-------|
| **Cursor** | `~/.cursor/mcp.json` | `mcpServers` | |
| **VS Code** | `~/.vscode/mcp.json` | `mcpServers` | Requires VS Code 1.99+ with MCP support |
| **Windsurf** | `~/.codeium/windsurf/mcp_config.json` | `mcpServers` | |
| **Zed** | `~/.config/zed/settings.json` | `context_servers` | Different schema; stdio only (http/sse servers skipped with a warning) |
| **Continue** | `~/.continue/config.json` | `mcpServers` | |
| **Junie** | `~/.junie/mcp/mcp.json` | `mcpServers` | JetBrains AI assistant |

### AI CLI Tools

| Agent | Config File | Key | Notes |
|-------|-------------|-----|-------|
| **Gemini CLI** | `~/.gemini/settings.json` | `mcpServers` | Google's Gemini CLI |
| **OpenAI Codex CLI** | `~/.codex/config.json` | `mcpServers` | OpenAI's open-source CLI |
| **GitHub Copilot CLI** | `~/.copilot/mcp-config.json` | `mcpServers` | Uses `type: "local"` instead of `"stdio"` — MCPSync normalizes automatically |
| **GitHub Copilot (VS Code)** | `~/.config/github-copilot/mcp.json` | `servers` | Different key name — MCPSync handles it |

### Don't see your agent?

- **Use `mcpsync add --path`** to register any custom config file right now
- **[Open an agent request issue](https://github.com/anushkrishnav/mcpsync/issues/new?template=agent-request.yml)** — include the config file path and a sample snippet if you have one
- **[Submit a PR](CONTRIBUTING.md)** — adding a new agent is usually ~10 lines of Go

---

## How It Works

```
Read  ──→  Normalize  ──→  Merge  ──→  Denormalize  ──→  Write
```

1. **Read** — each agent adapter reads its config file and normalizes servers to a common `NormalizedServer` struct (type, command, args, env, url, headers)
2. **Merge** — servers are deduplicated by name; when two agents have the same server name with different configs, the higher-priority agent wins (see priority order below)
3. **Write** — each adapter converts the merged list back into the agent's native format and writes it atomically (write to `.tmp` → rename)
4. **Backup** — before any write, the original is copied to `filename.old.json`

### Conflict resolution

If two agents have the same server name with different configurations, MCPSync pauses and shows you both versions side-by-side, then asks how to resolve it:

```
  ──────────────────────────────────────────────────────────────
  Conflict 1 of 1: "playwright"
  ──────────────────────────────────────────────────────────────

  [A]  Claude Desktop       (higher priority)
       type:      stdio
       command:   npx
       args:      @playwright/mcp@latest

  [B]  GitHub Copilot CLI   (lower priority)
       type:      stdio
       command:   npx
       args:      @playwright/mcp@latest
       extras:    tools

  ──────────────────────────────────────────────────────────────
  Conflict [1/1] — resolve "playwright":
  > Keep    "playwright" from Claude Desktop       stdio  npx @playwright/mcp@latest
    Use     "playwright" from GitHub Copilot CLI   stdio  npx @playwright/mcp@latest
    Rename  GitHub Copilot CLI's version and keep both
    Exclude "playwright" from this sync
    Delete  "playwright" from all agents
```

**Resolution options:**

| Option | What happens |
|--------|-------------|
| **Keep** | Use the higher-priority agent's version (A) |
| **Use other** | Replace with the lower-priority agent's version (B) |
| **Rename** | Keep A as-is, add B under a new name you provide — you get both |
| **Exclude** | Skip this server; it won't be written to any agent this sync |
| **Delete** | Remove from the merged list; effectively deleted from all synced agents |

Priority order (determines which version is [A]): Claude Desktop → Claude Code → Cursor → VS Code → Windsurf → Gemini → Copilot CLI → Copilot VS Code → Continue → Junie → Zed → Codex

### Schema normalization

Different agents use different type names and keys — MCPSync handles it transparently:

| Agent | Their format | MCPSync normalizes to |
|-------|--------------|-----------------------|
| GitHub Copilot CLI | `type: "local"` | `type: "stdio"` internally, written back as `"local"` |
| GitHub Copilot VS Code | key: `"servers"` | read/written as `"servers"`, stored as `NormalizedServer` |
| Zed | key: `"context_servers"`, nested `command.path` | transforms on read/write; skips http/sse |

### Claude Code project syncing

Claude Code stores MCP servers per-project inside `~/.claude.json`. MCPSync:
- **Reads** from all projects (takes the union of all unique server names)
- **Writes** the merged list back to all existing project entries

This means all your Claude Code projects get the same MCP server set after syncing.

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--maintain` | `true` | Back up original files as `.old.json` before writing |
| `--no-maintain` | — | Skip backups |
| `-p`, `--path <file>` | — | (`now` only) Also write merged JSON to this file |

---

## Citing this project

If you use MCPSync in research or reference it in a project, please cite it.
GitHub shows a **"Cite this repository"** button in the sidebar (powered by [`CITATION.cff`](CITATION.cff)).

**BibTeX:**
```bibtex
@software{mcpsync2026,
  author  = {anushkrishnav},
  title   = {{MCPSync}: Sync MCP Server Configurations Across AI Coding Agents},
  year    = {2026},
  version = {0.1.0},
  url     = {https://github.com/anushkrishnav/mcpsync},
  license = {MIT}
}
```

**APA:**
```
anushkrishnav. (2026). MCPSync (Version 0.1.0) [Computer software].
https://github.com/anushkrishnav/mcpsync
```

More formats (MLA, Chicago, IEEE, plain text): [`docs/citation.md`](docs/citation.md)

---

## License

[MIT](LICENSE) — free to use, modify, and distribute.

---

## Contributing

MCPSync is open source and community-driven. See [CONTRIBUTING.md](CONTRIBUTING.md).

- **Add an agent** — ~10 lines of Go, see the contributing guide
- **Report a bug** — [open a bug report](https://github.com/anushkrishnav/mcpsync/issues/new?template=bug_report.yml)
- **Request an agent** — [open an agent request](https://github.com/anushkrishnav/mcpsync/issues/new?template=agent-request.yml)
- **General ideas** — [open a discussion](https://github.com/anushkrishnav/mcpsync/discussions)
