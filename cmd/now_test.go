package cmd

import (
	"testing"

	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"github.com/anush-data-portfolio/MCPSync/internal/ui"
)

func TestApplyResolutions_OverwriteReplacesServer(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "old-cmd"},
	}
	newServer := agent.NormalizedServer{Name: "tool", Type: "stdio", Command: "new-cmd"}
	resolutions := []ui.ConflictResolution{
		{ServerName: "tool", Action: "overwrite", NewServer: newServer},
	}

	result := applyResolutions(servers, resolutions)
	if len(result) != 1 {
		t.Fatalf("expected 1 server, got %d", len(result))
	}
	if result[0].Command != "new-cmd" {
		t.Errorf("overwrite: expected 'new-cmd', got %q", result[0].Command)
	}
}

func TestApplyResolutions_ExcludeDropsServer(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "keep-me", Type: "stdio", Command: "cmd1"},
		{Name: "drop-me", Type: "stdio", Command: "cmd2"},
	}
	resolutions := []ui.ConflictResolution{
		{ServerName: "drop-me", Action: "exclude"},
	}

	result := applyResolutions(servers, resolutions)
	if len(result) != 1 {
		t.Fatalf("expected 1 server after exclude, got %d", len(result))
	}
	if result[0].Name != "keep-me" {
		t.Errorf("wrong server remained: %q", result[0].Name)
	}
}

func TestApplyResolutions_DeleteDropsServer(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "keep-me", Type: "stdio", Command: "cmd1"},
		{Name: "del-me", Type: "stdio", Command: "cmd2"},
	}
	resolutions := []ui.ConflictResolution{
		{ServerName: "del-me", Action: "delete"},
	}

	result := applyResolutions(servers, resolutions)
	if len(result) != 1 {
		t.Fatalf("expected 1 server after delete, got %d", len(result))
	}
	if result[0].Name != "keep-me" {
		t.Errorf("wrong server remained: %q", result[0].Name)
	}
}

func TestApplyResolutions_RenameAppendsNewServer(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "cmd1"},
	}
	renamedServer := agent.NormalizedServer{Name: "tool-renamed", Type: "stdio", Command: "cmd2"}
	resolutions := []ui.ConflictResolution{
		{ServerName: "tool", Action: "rename", NewServer: renamedServer},
	}

	result := applyResolutions(servers, resolutions)
	if len(result) != 2 {
		t.Fatalf("expected 2 servers after rename, got %d", len(result))
	}
	names := map[string]bool{}
	for _, s := range result {
		names[s.Name] = true
	}
	if !names["tool"] {
		t.Error("original 'tool' should still be present")
	}
	if !names["tool-renamed"] {
		t.Error("renamed 'tool-renamed' should be appended")
	}
}

func TestApplyResolutions_KeepChangesNothing(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "tool", Type: "stdio", Command: "original"},
	}
	resolutions := []ui.ConflictResolution{
		{ServerName: "tool", Action: "keep"},
	}

	result := applyResolutions(servers, resolutions)
	if len(result) != 1 {
		t.Fatalf("expected 1 server, got %d", len(result))
	}
	if result[0].Command != "original" {
		t.Errorf("keep: command changed to %q", result[0].Command)
	}
}

func TestApplyResolutions_SortedAlphabetically(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "zebra", Type: "stdio", Command: "z"},
		{Name: "apple", Type: "stdio", Command: "a"},
		{Name: "mango", Type: "stdio", Command: "m"},
	}

	result := applyResolutions(servers, nil)
	if len(result) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(result))
	}
	if result[0].Name != "apple" || result[1].Name != "mango" || result[2].Name != "zebra" {
		t.Errorf("servers not sorted: %v", []string{result[0].Name, result[1].Name, result[2].Name})
	}
}

func TestApplyResolutions_NoResolutions(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "a", Type: "stdio", Command: "cmd-a"},
		{Name: "b", Type: "stdio", Command: "cmd-b"},
	}

	result := applyResolutions(servers, nil)
	if len(result) != 2 {
		t.Fatalf("expected 2 servers with no resolutions, got %d", len(result))
	}
}

func TestApplyResolutions_MultipleActions(t *testing.T) {
	servers := []agent.NormalizedServer{
		{Name: "keep", Type: "stdio", Command: "k"},
		{Name: "overwrite-me", Type: "stdio", Command: "old"},
		{Name: "exclude-me", Type: "stdio", Command: "e"},
	}
	newServer := agent.NormalizedServer{Name: "overwrite-me", Type: "stdio", Command: "new"}
	resolutions := []ui.ConflictResolution{
		{ServerName: "overwrite-me", Action: "overwrite", NewServer: newServer},
		{ServerName: "exclude-me", Action: "exclude"},
	}

	result := applyResolutions(servers, resolutions)
	if len(result) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(result))
	}
	names := map[string]string{}
	for _, s := range result {
		names[s.Name] = s.Command
	}
	if names["keep"] != "k" {
		t.Error("'keep' server should be unchanged")
	}
	if names["overwrite-me"] != "new" {
		t.Errorf("'overwrite-me' should have new command, got %q", names["overwrite-me"])
	}
	if _, ok := names["exclude-me"]; ok {
		t.Error("'exclude-me' should have been dropped")
	}
}
