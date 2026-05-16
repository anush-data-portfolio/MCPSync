# Contributing to MCPSync

Thank you for contributing! MCPSync is community-driven — most of the value comes from broad agent coverage.

## Ways to contribute

- **Add a new agent** — the most common contribution (see below)
- **Fix a bug** — open a bug report issue, then submit a PR
- **Improve docs** — clarify README, add examples, fix typos
- **Test on your platform** — verify macOS/Linux/Windows behavior

---

## Adding a new agent

Most agents take about 10 lines of Go. Here's how:

### 1. Create `internal/agent/<agentname>.go`

**Standard case** (flat JSON with `mcpServers` or `servers` key):

```go
package agent

func init() {
    Register(&StandardAgent{
        id:          "my-agent",
        displayName: "My Agent",
        defaultPath: "~/.my-agent/mcp.json",
        serverKey:   "mcpServers",
    })
}
```

That's it. The `StandardAgent` base class handles read/write/backup automatically.

**Non-standard case** (different JSON structure, type mapping, etc.):

Look at `claude_code.go` (per-project nesting), `copilot_cli.go` (`type: "local"` mapping), or `zed.go` (`context_servers` format) as examples. Implement the `Agent` interface directly.

### 2. Add the agent to the bug report template

In `.github/ISSUE_TEMPLATE/bug_report.yml`, add the agent name to the `affected_agent` options list.

### 3. Update the README agent table

In `README.md`, add a row to the **Supported Agents** table.

### 4. Update the CHANGELOG

Add an entry under the next version in `CHANGELOG.md`.

---

## Agent interface

```go
type Agent interface {
    ID() string                  // unique lowercase-kebab-case ID: "my-agent"
    DisplayName() string         // human label: "My Agent"
    DefaultConfigPath() string   // resolved absolute path (use resolvePath("~/..."))
    ResolvedConfigPath(override string) string
    IsPresent(configPath string) bool
    Read(configPath string) (*AgentReadResult, error)
    Write(servers []NormalizedServer, configPath string, rawConfig map[string]any, maintain bool) error
}
```

`Read()` must return `nil, nil` (not an error) if the config file simply doesn't exist — that means "agent not installed."

`Write()` receives `rawConfig` (the original parsed JSON from `Read()`) so it can preserve all non-MCP keys in the file. It should call `writeJSON()` for the actual file write, which handles atomic writes and `.old.json` backups.

---

## Building locally

```bash
git clone https://github.com/anushkrishnav/mcpsync
cd mcpsync
go mod tidy
go build -o mcpsync .
./mcpsync list   # smoke test
```

Requires Go 1.21+. If you only want to test the npm shim flow without publishing:

```bash
# Simulate what npm install -g mcpsync does
go build -o bin/mcpsync .     # put binary where the JS shim expects it
node bin/mcpsync.js list      # run via the JS wrapper
```

---

## Pull request checklist

- [ ] New agent file in `internal/agent/`
- [ ] Agent added to bug report template
- [ ] README agent table updated
- [ ] CHANGELOG entry added
- [ ] `go build ./...` passes

No tests are required for simple `StandardAgent` additions. For complex adapters, a brief mention in the PR description of how you verified it works (what config file you tested against) is enough.

---

## Finding MCP config paths

If you're unsure where an agent stores its MCP config:

```bash
# Search home directory for JSON files that look like MCP configs
find ~ -maxdepth 5 -name "*.json" 2>/dev/null | xargs grep -l "mcpServers" 2>/dev/null

# Check common agent directories
ls -la ~/.cursor ~/.vscode ~/.codeium ~/.gemini ~/.codex 2>/dev/null
```

On macOS:
```bash
ls ~/Library/Application\ Support/ | grep -i "claude\|cursor\|wind\|gemini\|copilot"
```
