# Commet

**Automated semantic versioning based on commit messages**

Commet analyzes your git commit history and automatically updates version numbers in your project files based on conventional commit messages.

## Features

- 🚀 Automatic semantic version bumping based on commit types
- 📦 Support for JSON (composer.json, package.json) and YAML (config.yaml) files
- 🎯 Configurable commit type to version bump mapping
- 🏷️ Git tag-based and file-based version detection
- 🔧 Dry-run mode to preview changes
- 🎨 Colored output for better readability
- 🤖 Optional auto-commit and auto-tag
- 📝 Multiple version file support

### Supported Formats

1. **Type with scope**: `Feature(log): added logger wrap`
2. **Type without scope**: `Fix: handle null responses`
3. **Board with wrapped type**: `J-123456(parser,regex): <Fix> syntax issue`
4. **Board with unwrapped type**: `U-1234(config): Feature new section`
5. **Force major**: `Fix!(core): Removed endpoint` or `Breaking: change`

## Installation

### From Source

```bash
git clone https://github.com/yendefrr/commet.git
cd commet
go build -o commet cmd/commet/main.go
sudo mv commet /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yendefrr/commet/cmd/commet@latest
```

## Quick Start

1. Initialize a config file in your project:

```bash
commet init
```

2. Edit `.commet.toml` to match your project structure


3. Run commet:

```bash
# Preview changes
commet --dry-run

# Apply version update
commet

# Verbose output
commet --verbose

# Commit version update (if disabled auto)
commet commit

# Commit version update with tag (if disabled auto)
commet commit --tag
```

## Configuration

### Basic Configuration

```toml
[version]
file = "config.yaml"    # Path to version file
key = "app.version"     # Key path (dot notation for nested)
initial = "0.1.0"       # Initial version if none exists
format = "semver"       # "semver" (1.2.3) or "v-prefix" (v1.2.3)

[bump_rules]
Fix = "patch"        # Bug fixes
Feature = "minor"    # New features
Refactor = "patch"   # Code refactoring
Breaking = "major"   # Breaking changes
"!" = "major"        # Force major (Type!)
Docs = "none"        # No version bump
Tests = "none"       # No version bump
Style = "none"       # No version bump

[detection]
strategies = ["git-tags", "version-file"]  # Detect from git tags, then version file
tag_pattern = '^v?([0-9]+\.[0-9]+\.[0-9]+)$'
exclude_merges = true

# Git operations
[git]
auto_commit = false
commit_message = "Conf: bump version to {version}"
auto_tag = false
tag_format = "v{version}"
tag_message = "Release {version}"

# Multiple version files
[[additional_files]]
file = "package.json"
key = "version"

[[additional_files]]
file = "Chart.yaml"
key = "version"
```

## CLI Usage

```bash
Usage:
  commet [flags]
  commet [command]

Available Commands:
  commit      Commit version changes to git
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize a new .commet.toml configuration file

Flags:
      --config string   config file (default is .commet.toml)
      --dry-run         show what would be done without making changes
      --from string     start ref for commit range
  -h, --help            help for commet
      --to string       end ref for commit range (default "HEAD")
      --verbose         verbose output

Use "commet [command] --help" for more information about a command.
```

### Precedence

When multiple commits are present, the highest bump type take priority:

```
MAJOR > MINOR > PATCH > NONE
```

### Examples

Current version: `1.2.3`

```bash
# Commits:
# - Fix(api): handle errors
# Result: 1.2.4 (PATCH)

# Commits:
# - Feature(auth): add OAuth
# - Fix(api): handle errors
# Result: 1.3.0 (MINOR - higher precedence)

# Commits:
# - Fix!(security): critical fix
# - Feature(auth): add OAuth
# Result: 2.0.0 (MAJOR - highest precedence)
```

## Version Detection

Commet uses multiple strategies to detect the current version:

1. **Git tags** (default): Finds the latest git tag matching the pattern
2. **Version file**: Reads version from the configured file
3. **Initial version**: Falls back to configured initial version (default: 0.1.0)

You can configure the detection order in `.commet.yaml`:

```yaml
detection:
  strategies:
    - git-tags
    - version-file
```


## Related Project

Works perfectly with [meteor](https://github.com/stefanlogue/meteor)
