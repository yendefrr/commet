package changelog

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/yendefrr/commet/internal/parser"
)

type Generator struct {
	filePath string
}

func NewGenerator(filePath string) *Generator {
	return &Generator{filePath: filePath}
}

type CommitGroup struct {
	Type        string
	Emoji       string
	Description string
	Commits     []*parser.Commit
}

func (g *Generator) Generate(version string, commits []*parser.Commit) error {
	// Group commits by type
	groups := g.groupCommits(commits)

	// Generate markdown
	entry := g.formatEntry(version, groups)

	// Append to file
	return g.appendToFile(entry)
}

func (g *Generator) groupCommits(commits []*parser.Commit) []*CommitGroup {
	typeMap := make(map[string]*CommitGroup)
	var untyped []*parser.Commit

	// Define type metadata
	typeMetadata := map[string]struct {
		emoji       string
		description string
	}{
		"Feature":   {"✨", "Features"},
		"Fix":       {"🐝", "Bug Fixes"},
		"Refactor":  {"🔧", "Refactoring"},
		"Docs":      {"📚", "Documentation"},
		"Style":     {"💅", "Styling"},
		"Build":     {"🏗️", "Build System"},
		"Tests":     {"🧪", "Tests"},
		"Conf":      {"🧰", "Configuration"},
		"Migrations": {"🗄️", "Migrations"},
		"Submodule": {"🏷️", "Submodules"},
		"Breaking":  {"💥", "Breaking Changes"},
	}

	for _, commit := range commits {
		if commit.Type == "" {
			untyped = append(untyped, commit)
			continue
		}

		if _, exists := typeMap[commit.Type]; !exists {
			meta := typeMetadata[commit.Type]
			typeMap[commit.Type] = &CommitGroup{
				Type:        commit.Type,
				Emoji:       meta.emoji,
				Description: meta.description,
				Commits:     []*parser.Commit{},
			}
			if typeMap[commit.Type].Description == "" {
				typeMap[commit.Type].Description = commit.Type
			}
		}

		typeMap[commit.Type].Commits = append(typeMap[commit.Type].Commits, commit)
	}

	var groups []*CommitGroup

	typeOrder := []string{
		"Breaking",
		"Feature",
		"Fix",
		"Refactor",
		"Docs",
		"Style",
		"Build",
		"Tests",
		"Conf",
		"Migrations",
		"Submodule",
	}

	for _, typeName := range typeOrder {
		if group, exists := typeMap[typeName]; exists {
			groups = append(groups, group)
			delete(typeMap, typeName)
		}
	}

	for _, group := range typeMap {
		groups = append(groups, group)
	}

	if len(untyped) > 0 {
		groups = append(groups, &CommitGroup{
			Type:        "",
			Emoji:       "📝",
			Description: "Other Changes",
			Commits:     untyped,
		})
	}

	return groups
}

func (g *Generator) formatEntry(version string, groups []*CommitGroup) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## [%s] - %s\n\n", version, time.Now().Format("2006-01-02")))

	for _, group := range groups {
		if len(group.Commits) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s %s\n\n", group.Emoji, group.Description))

		for _, commit := range group.Commits {
			sb.WriteString(g.formatCommit(commit))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *Generator) formatCommit(commit *parser.Commit) string {
	var parts []string

	if commit.Scope != "" {
		parts = append(parts, fmt.Sprintf("**%s**", commit.Scope))
	}

	parts = append(parts, commit.Description)

	var suffix string
	if commit.Board != "" {
		suffix = fmt.Sprintf(" (%s)", commit.Board)
	}

	if commit.Hash != "" {
		suffix += fmt.Sprintf(" [`%s`]", commit.Hash)
	}

	return fmt.Sprintf("- %s%s\n", strings.Join(parts, ": "), suffix)
}

func (g *Generator) appendToFile(entry string) error {
	var content []byte

	if _, err := os.Stat(g.filePath); err == nil {
		var readErr error
		content, readErr = os.ReadFile(g.filePath)
		if readErr != nil {
			return fmt.Errorf("failed to read changelog: %w", readErr)
		}
	} else {
		content = []byte("# Changelog\n\nAll notable changes to this project will be documented in this file.\n\n")
	}

	lines := strings.Split(string(content), "\n")
	var newContent strings.Builder

	headerEnd := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "## [") {
			headerEnd = i
			break
		}
		if i < len(lines)-1 {
			newContent.WriteString(line + "\n")
		}
	}

	if headerEnd == 0 {
		headerEnd = len(lines)
		newContent.WriteString("\n")
	}

	newContent.WriteString(entry)

	for i := headerEnd; i < len(lines); i++ {
		if i == len(lines)-1 && lines[i] == "" {
			continue
		}
		newContent.WriteString(lines[i] + "\n")
	}

	if err := os.WriteFile(g.filePath, []byte(newContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}

func GetCommitsSinceVersion(commits []*parser.Commit, version string) []*parser.Commit {
	return commits
}

func SortCommits(commits []*parser.Commit) {
	typePriority := map[string]int{
		"Breaking":  1,
		"Feature":   2,
		"Fix":       3,
		"Refactor":  4,
		"Docs":      5,
		"Style":     6,
		"Build":     7,
		"Tests":     8,
		"Conf":      9,
		"Migrations": 10,
		"Submodule": 11,
		"":          99,
	}

	sort.Slice(commits, func(i, j int) bool {
		priI := typePriority[commits[i].Type]
		priJ := typePriority[commits[j].Type]
		return priI < priJ
	})
}
