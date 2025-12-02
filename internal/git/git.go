package git

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/yendefrr/commet/internal/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Client struct {
	repo   *git.Repository
	config *config.Config
}

func NewClient(repoPath string, cfg *config.Config) (*Client, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &Client{
		repo:   repo,
		config: cfg,
	}, nil
}

type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

func (c *Client) GetCommits(from, to string) ([]*CommitInfo, error) {
	if from == "" {
		latestTag, err := c.GetLatestTag()
		if err == nil && latestTag != "" {
			from = latestTag
		}
	}

	toRef, err := c.repo.ResolveRevision(plumbing.Revision(to))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'to' ref %s: %w", to, err)
	}

	logOptions := &git.LogOptions{
		From: *toRef,
	}

	commitIter, err := c.repo.Log(logOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}
	defer commitIter.Close()

	var fromHash plumbing.Hash
	if from != "" {
		fromRef, err := c.repo.ResolveRevision(plumbing.Revision(from))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve 'from' ref %s: %w", from, err)
		}
		fromHash = *fromRef
	}

	var commits []*CommitInfo
	err = commitIter.ForEach(func(commit *object.Commit) error {
		if from != "" && commit.Hash == fromHash {
			return fmt.Errorf("stop")
		}

		if c.config.Detection.ExcludeMerges && len(commit.ParentHashes) > 1 {
			return nil
		}

		message := strings.Split(commit.Message, "\n")[0]

		commits = append(commits, &CommitInfo{
			Hash:    commit.Hash.String()[:7],
			Message: message,
			Author:  commit.Author.Name,
			Date:    commit.Author.When.Format("2006-01-02"),
		})

		return nil
	})

	if err != nil && err.Error() != "stop" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

func (c *Client) GetLatestTag() (string, error) {
	tags, err := c.repo.Tags()
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}
	defer tags.Close()

	pattern, err := regexp.Compile(c.config.Detection.TagPattern)
	if err != nil {
		return "", fmt.Errorf("invalid tag pattern: %w", err)
	}

	var matchingTags []string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagName := ref.Name().Short()
		if pattern.MatchString(tagName) {
			matchingTags = append(matchingTags, tagName)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(matchingTags) == 0 {
		return "", nil
	}

	sort.Slice(matchingTags, func(i, j int) bool {
		return matchingTags[i] > matchingTags[j]
	})

	return matchingTags[0], nil
}

func (c *Client) ExtractVersionFromTag(tag string) (string, error) {
	pattern, err := regexp.Compile(c.config.Detection.TagPattern)
	if err != nil {
		return "", fmt.Errorf("invalid tag pattern: %w", err)
	}

	matches := pattern.FindStringSubmatch(tag)
	if len(matches) < 2 {
		return "", fmt.Errorf("tag does not match pattern")
	}

	return matches[1], nil
}

func (c *Client) CreateCommit(files []string, message string) error {
	worktree, err := c.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	for _, file := range files {
		_, err := worktree.Add(file)
		if err != nil {
			return fmt.Errorf("failed to add file %s: %w", file, err)
		}
	}

	_, err = worktree.Commit(message, &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

func (c *Client) CreateTag(tag, message string) error {
	head, err := c.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	opts := &git.CreateTagOptions{
		Message: message,
	}

	_, err = c.repo.CreateTag(tag, head.Hash(), opts)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

func IsGitRepository(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}
