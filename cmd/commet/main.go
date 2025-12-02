package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yendefrr/commet/internal/config"
	"github.com/yendefrr/commet/internal/git"
	"github.com/yendefrr/commet/internal/parser"
	"github.com/yendefrr/commet/internal/updater"
	"github.com/yendefrr/commet/internal/version"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	dryRun  bool
	verbose bool
	fromRef string
	toRef   string
)

var rootCmd = &cobra.Command{
	Use:   "commet",
	Short: "Automated semantic versioning based on commits",
	Long: `Commet analyzes your commit history and automatically
updates version numbers in your project files based on
conventional commit messages.`,
	RunE: run,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new .commet.toml configuration file",
	Long:  `Creates a new .commet.toml file with default configuration in the current directory.`,
	RunE:  initConfig,
}

func init() {
	rootCmd.AddCommand(initCmd)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .commet.toml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&fromRef, "from", "", "start ref for commit range")
	rootCmd.PersistentFlags().StringVar(&toRef, "to", "HEAD", "end ref for commit range")
}

func initConfig(cmd *cobra.Command, args []string) error {
	configPath := ".commet.toml"

	if fileExists(configPath) {
		color.Yellow("Config file %s already exists", configPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			color.Cyan("Operation cancelled")
			return nil
		}
	}

	cfg := config.DefaultConfig()

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	color.Green("✓ Created %s", configPath)
	fmt.Println()
	color.Cyan("Next steps:")
	fmt.Println("  1. Edit .commet.toml to match your project")
	fmt.Println("  2. Run 'commet --dry-run' to preview changes")
	fmt.Println("  3. Run 'commet' to bump version")

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if verbose {
		color.Cyan("[CONFIG] Loaded configuration")
		if cfgFile != "" {
			color.Cyan("[CONFIG] File: %s", cfgFile)
		}
	}

	// Check git repository
	if !git.IsGitRepository(".") {
		return fmt.Errorf("not a git repository")
	}

	// Create git client
	gitClient, err := git.NewClient(".", cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	// Detect current version
	currentVersion, err := detectVersion(gitClient, cfg)
	if err != nil {
		return fmt.Errorf("failed to detect current version: %w", err)
	}

	if verbose {
		color.Cyan("[VERSION] Current: %s", currentVersion)
	}

	// Get commits
	commits, err := gitClient.GetCommits(fromRef, toRef)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		color.Yellow("No commits found since %s", currentVersion)
		return nil
	}

	if verbose {
		color.Cyan("[GIT] Found %d commits", len(commits))
	}

	// Parse commits
	parsedCommits := make([]*parser.Commit, 0, len(commits))
	for _, c := range commits {
		parsed, err := parser.Parse(c.Message)
		if err != nil {
			if verbose {
				color.Yellow("[WARN] Failed to parse: %s", c.Message)
			}
			continue
		}

		if !parsed.IsValidCommit() {
			if verbose {
				color.Yellow("[WARN] Invalid commit format: %s", c.Message)
			}
			continue
		}

		parsed.Hash = c.Hash
		parsedCommits = append(parsedCommits, parsed)

		if verbose {
			bump := cfg.GetBumpType(parsed.Type)
			forceMark := ""
			if parsed.ForceMajor {
				forceMark = " [FORCE MAJOR]"
			}
			fmt.Printf("  %s → %s%s\n", truncate(c.Message, 60), bump, forceMark)
		}
	}

	if len(parsedCommits) == 0 {
		color.Yellow("No valid commits found")
		return nil
	}

	// Calculate new version
	calculator := version.NewCalculator(cfg)
	newVersion, bumpType, err := calculator.Calculate(currentVersion, parsedCommits)
	if err != nil {
		return fmt.Errorf("failed to calculate version: %w", err)
	}

	if bumpType == config.BumpNone {
		color.Green("No version bump needed (current: %s)", currentVersion)
		return nil
	}

	// Display results
	fmt.Println()
	color.Green("Current version: %s", currentVersion)
	color.Green("Next version:    %s", newVersion)
	color.Green("Bump type:       %s", strings.ToUpper(string(bumpType)))
	fmt.Println()

	if dryRun {
		color.Yellow("Files to update:")
		for _, versionFile := range cfg.GetVersionFiles() {
			color.Yellow("  - %s (%s)", versionFile.File, versionFile.Key)
		}
		fmt.Println()
		color.Yellow("No changes made (dry run mode)")
		return nil
	}

	// Update version files
	updatedFiles := []string{}
	for _, versionFile := range cfg.GetVersionFiles() {
		filePath := versionFile.File
		if !fileExists(filePath) {
			color.Yellow("[WARN] File not found: %s", filePath)
			continue
		}

		fileUpdater, err := updater.New(filePath)
		if err != nil {
			return fmt.Errorf("failed to create updater for %s: %w", filePath, err)
		}

		if err := fileUpdater.SetVersion(versionFile.Key, newVersion); err != nil {
			return fmt.Errorf("failed to update %s: %w", filePath, err)
		}

		color.Green("✓ Updated %s", filePath)
		updatedFiles = append(updatedFiles, filePath)
	}

	// Git operations
	if cfg.Git.AutoCommit && len(updatedFiles) > 0 {
		commitMsg := strings.ReplaceAll(cfg.Git.CommitMessage, "{version}", newVersion)
		if err := gitClient.CreateCommit(updatedFiles, commitMsg); err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}
		color.Green("✓ Created commit: %s", commitMsg)
	}

	if cfg.Git.AutoTag {
		tagName := strings.ReplaceAll(cfg.Git.TagFormat, "{version}", newVersion)
		tagMsg := strings.ReplaceAll(cfg.Git.TagMessage, "{version}", newVersion)
		if err := gitClient.CreateTag(tagName, tagMsg); err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}
		color.Green("✓ Created tag: %s", tagName)
	}

	fmt.Println()
	color.Green("Version updated: %s → %s", currentVersion, newVersion)

	return nil
}

func detectVersion(gitClient *git.Client, cfg *config.Config) (string, error) {
	for _, strategy := range cfg.Detection.Strategies {
		switch strategy {
		case "git-tags":
			tag, err := gitClient.GetLatestTag()
			if err == nil && tag != "" {
				version, err := gitClient.ExtractVersionFromTag(tag)
				if err == nil {
					return version, nil
				}
			}

		case "version-file":
			filePath := cfg.Version.File
			if fileExists(filePath) {
				fileUpdater, err := updater.New(filePath)
				if err == nil {
					version, err := fileUpdater.GetVersion(cfg.Version.Key)
					if err == nil && version != "" {
						return version, nil
					}
				}
			}
		}
	}

	return cfg.Version.Initial, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
