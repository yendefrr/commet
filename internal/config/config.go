package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Version         VersionConfig       `toml:"version"`
	BumpRules       map[string]BumpType `toml:"bump_rules"`
	Detection       DetectionConfig     `toml:"detection"`
	Git             GitConfig           `toml:"git"`
	Changelog       ChangelogConfig     `toml:"changelog"`
	AdditionalFiles []VersionConfig     `toml:"additional_files,omitempty"`
}

type VersionConfig struct {
	File    string `toml:"file"`
	Key     string `toml:"key"`
	Initial string `toml:"initial"`
	Format  string `toml:"format"` // "semver" or "v-prefix"
}

type BumpType string

const (
	BumpNone  BumpType = "none"
	BumpPatch BumpType = "patch"
	BumpMinor BumpType = "minor"
	BumpMajor BumpType = "major"
)

type DetectionConfig struct {
	Strategies    []string `toml:"strategies"`
	TagPattern    string   `toml:"tag_pattern"`
	ExcludeMerges bool     `toml:"exclude_merges"`
}

type GitConfig struct {
	AutoCommit    bool   `toml:"auto_commit"`
	CommitMessage string `toml:"commit_message"`
	AutoTag       bool   `toml:"auto_tag"`
	TagFormat     string `toml:"tag_format"`
	TagMessage    string `toml:"tag_message"`
}

type ChangelogConfig struct {
	Enabled bool   `toml:"enabled"`
	File    string `toml:"file"`
}

func DefaultConfig() *Config {
	return &Config{
		Version: VersionConfig{
			File:    "composer.json",
			Key:     "version",
			Initial: "0.1.0",
			Format:  "semver",
		},
		BumpRules: map[string]BumpType{
			"Fix":      BumpPatch,
			"Feature":  BumpMinor,
			"Refactor": BumpPatch,
			"Style":    BumpNone,
			"Docs":     BumpNone,
			"Build":    BumpPatch,
			"Tests":    BumpNone,
			"Breaking": BumpMajor,
			"!":        BumpMajor,
		},
		Detection: DetectionConfig{
			Strategies:    []string{"git-tags", "version-file"},
			TagPattern:    `^v?([0-9]+\.[0-9]+\.[0-9]+)$`,
			ExcludeMerges: true,
		},
		Git: GitConfig{
			AutoCommit:    false,
			CommitMessage: "Conf: bump version to {version}",
			AutoTag:       false,
			TagFormat:     "v{version}",
			TagMessage:    "Release {version}",
		},
		Changelog: ChangelogConfig{
			Enabled: false,
			File:    "CHANGELOG.md",
		},
	}
}

func Load(configPath string) (*Config, error) {
	if configPath == "" {
		if _, err := os.Stat(".commet.toml"); err == nil {
			configPath = ".commet.toml"
		} else {
			return DefaultConfig(), nil
		}
	}

	cfg := DefaultConfig()
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Version.File == "" {
		return fmt.Errorf("version.file is required")
	}

	if c.Version.Key == "" {
		return fmt.Errorf("version.key is required")
	}

	if c.Version.Initial == "" {
		c.Version.Initial = "0.1.0"
	}

	if c.Version.Format == "" {
		c.Version.Format = "semver"
	}
	if c.Version.Format != "semver" && c.Version.Format != "v-prefix" {
		return fmt.Errorf("version.format must be 'semver' or 'v-prefix'")
	}

	if len(c.BumpRules) == 0 {
		return fmt.Errorf("bump_rules cannot be empty")
	}

	if len(c.Detection.Strategies) == 0 {
		c.Detection.Strategies = []string{"git-tags", "version-file"}
	}

	if c.Detection.TagPattern == "" {
		c.Detection.TagPattern = `^v?([0-9]+\.[0-9]+\.[0-9]+)$`
	}

	return nil
}

func (c *Config) GetBumpType(commitType string) BumpType {
	if bump, ok := c.BumpRules[commitType]; ok {
		return bump
	}
	return BumpNone
}

func (c *Config) GetVersionFiles() []VersionConfig {
	files := []VersionConfig{c.Version}
	files = append(files, c.AdditionalFiles...)
	return files
}

func (c *Config) ResolveVersionFilePath(configPath string) string {
	if filepath.IsAbs(c.Version.File) {
		return c.Version.File
	}

	if configPath != "" {
		configDir := filepath.Dir(configPath)
		return filepath.Join(configDir, c.Version.File)
	}

	return c.Version.File
}

func (c *Config) Save(configPath string) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
