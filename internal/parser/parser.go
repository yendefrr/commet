package parser

import (
	"regexp"
	"strings"
)

type Commit struct {
	Hash        string
	Message     string
	Type        string
	Scope       string
	Board       string
	Description string
	ForceMajor  bool
}

var (
	// Pattern 1: J-123456(parser,regex): <Fix> syntax issue
	pattern1 = regexp.MustCompile(`^(?P<board>[A-Z]+-\d+)(?:\((?P<scope>[^)]+)\))?: <(?P<type>[^>]+)>\s*(?P<desc>.+)$`)

	// Pattern 2: U-1234(config): Feature new section
	pattern2 = regexp.MustCompile(`^(?P<board>[A-Z]+-\d+)\((?P<scope>[^)]+)\): (?P<type>\w+)\s+(?P<desc>.+)$`)

	// Pattern 3: U-1234: Tests added for parser
	pattern3 = regexp.MustCompile(`^(?P<board>[A-Z]+-\d+): (?P<type>\w+)\s+(?P<desc>.+)$`)

	// Pattern 4: Feature!(log): added logger
	pattern4 = regexp.MustCompile(`^(?P<type>\w+)(?P<force>!)?(?:\((?P<scope>[^)]+)\))?: (?P<desc>.+)$`)

	// All patterns in order of priority
	patterns = []*regexp.Regexp{pattern1, pattern2, pattern3, pattern4}
)

func Parse(message string) (*Commit, error) {
	commit := &Commit{
		Message: message,
	}

	message = strings.TrimSpace(message)

	if strings.Contains(message, "Breaking") || strings.Contains(message, "BREAKING") {
		commit.ForceMajor = true
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(message); matches != nil {
			names := pattern.SubexpNames()
			for i, name := range names {
				if i == 0 || name == "" {
					continue
				}
				value := matches[i]
				switch name {
				case "type":
					commit.Type = strings.TrimSpace(value)
				case "scope":
					commit.Scope = value
				case "board":
					commit.Board = value
				case "desc":
					commit.Description = value
				case "force":
					if value == "!" {
						commit.ForceMajor = true
					}
				}
			}
			return commit, nil
		}
	}


	// Fallback: try to parse type and description only
	parts := strings.SplitN(message, ":", 2)
	if len(parts) == 2 {
		possibleType := strings.TrimSpace(parts[0])
		if strings.HasSuffix(possibleType, "!") {
			commit.ForceMajor = true
			possibleType = strings.TrimSuffix(possibleType, "!")
		}
		commit.Type = possibleType
		commit.Description = strings.TrimSpace(parts[1])
		return commit, nil
	}

	return commit, nil
}

func ParseMultiple(messages []string) []*Commit {
	commits := make([]*Commit, 0, len(messages))
	for _, msg := range messages {
		if commit, err := Parse(msg); err == nil && commit.Type != "" {
			commits = append(commits, commit)
		}
	}
	return commits
}

func (c *Commit) IsValidCommit() bool {
	return c.Type != ""
}

func (c *Commit) String() string {
	var parts []string

	if c.Board != "" {
		parts = append(parts, c.Board)
	}

	if c.Type != "" {
		typeStr := c.Type
		if c.ForceMajor {
			typeStr += "!"
		}
		parts = append(parts, typeStr)
	}

	if c.Scope != "" {
		parts = append(parts, "("+c.Scope+")")
	}

	if c.Description != "" {
		parts = append(parts, c.Description)
	}

	return strings.Join(parts, " ")
}
