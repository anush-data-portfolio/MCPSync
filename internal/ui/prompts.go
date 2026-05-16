package ui

import (
	"fmt"
	"github.com/anush-data-portfolio/MCPSync/internal/agent"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

type ConflictResolution struct {
	ServerName string
	Action     string // "keep", "overwrite", "rename", "exclude", "delete"
	NewServer  agent.NormalizedServer
}

func ResolveConflict(conflict agent.ConflictRecord, idx, total int) (ConflictResolution, error) {
	PrintConflictDetail(conflict, idx, total)

	kept := conflict.Kept
	disc := conflict.Discarded

	options := []string{
		fmt.Sprintf("Keep    %q from %-20s  %s", conflict.ServerName, kept.SourceAgent, dim.Sprint(summarizeServer(kept))),
		fmt.Sprintf("Use     %q from %-20s  %s", conflict.ServerName, disc.SourceAgent, dim.Sprint(summarizeServer(disc))),
		fmt.Sprintf("Rename  %s's version and keep both", disc.SourceAgent),
		fmt.Sprintf("Exclude %q from this sync", conflict.ServerName),
		fmt.Sprintf("Delete  %q from all agents", conflict.ServerName),
	}

	var choice string
	prompt := &survey.Select{
		Message: fmt.Sprintf("Conflict [%d/%d] — resolve %q:", idx, total, conflict.ServerName),
		Options: options,
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return ConflictResolution{}, err
	}

	switch {
	case strings.HasPrefix(choice, "Keep"):
		return ConflictResolution{
			ServerName: conflict.ServerName,
			Action:     "keep",
		}, nil

	case strings.HasPrefix(choice, "Use"):
		winner := disc
		winner.Name = conflict.ServerName
		return ConflictResolution{
			ServerName: conflict.ServerName,
			Action:     "overwrite",
			NewServer:  winner,
		}, nil

	case strings.HasPrefix(choice, "Rename"):
		defaultName := conflict.ServerName + "-" + strings.ReplaceAll(disc.SourceAgent, " ", "-")
		newName, err := AskInput(
			fmt.Sprintf("New name for %s's %q:", disc.SourceAgent, conflict.ServerName),
			defaultName,
		)
		if err != nil {
			return ConflictResolution{}, err
		}
		renamed := disc
		renamed.Name = newName
		return ConflictResolution{
			ServerName: conflict.ServerName,
			Action:     "rename",
			NewServer:  renamed,
		}, nil

	case strings.HasPrefix(choice, "Exclude"):
		return ConflictResolution{
			ServerName: conflict.ServerName,
			Action:     "exclude",
		}, nil

	case strings.HasPrefix(choice, "Delete"):
		return ConflictResolution{
			ServerName: conflict.ServerName,
			Action:     "delete",
		}, nil
	}

	return ConflictResolution{ServerName: conflict.ServerName, Action: "keep"}, nil
}

func summarizeServer(s agent.NormalizedServer) string {
	switch s.Type {
	case "stdio":
		args := strings.Join(s.Args, " ")
		if args != "" {
			return fmt.Sprintf("stdio  %s %s", s.Command, args)
		}
		return fmt.Sprintf("stdio  %s", s.Command)
	case "http", "sse":
		url := s.URL
		if len(url) > 50 {
			url = url[:47] + "..."
		}
		return fmt.Sprintf("%s  %s", s.Type, url)
	}
	return s.Type
}

func SelectAgents(infos []agent.AgentInfo) ([]string, error) {
	options := make([]string, len(infos))
	defaults := []string{}

	for i, info := range infos {
		options[i] = info.DisplayName
		if info.Detected {
			defaults = append(defaults, info.DisplayName)
		}
	}

	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select agents to sync:",
		Options: options,
		Default: defaults,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, err
	}

	nameToID := make(map[string]string, len(infos))
	for _, info := range infos {
		nameToID[info.DisplayName] = info.ID
	}

	ids := make([]string, 0, len(selected))
	for _, name := range selected {
		if id, ok := nameToID[name]; ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func Confirm(message string) (bool, error) {
	var ok bool
	prompt := &survey.Confirm{Message: message, Default: true}
	if err := survey.AskOne(prompt, &ok); err != nil {
		return false, err
	}
	return ok, nil
}

func AskInput(message, defaultVal string) (string, error) {
	var answer string
	prompt := &survey.Input{Message: message, Default: defaultVal}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}
	return answer, nil
}

func SelectOne(message string, options []string) (string, error) {
	var answer string
	prompt := &survey.Select{Message: message, Options: options}
	if err := survey.AskOne(prompt, &answer); err != nil {
		return "", err
	}
	return answer, nil
}
