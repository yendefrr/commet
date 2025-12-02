package version

import (
	"fmt"
	"strings"

	"github.com/yendefrr/commet/internal/config"
	"github.com/yendefrr/commet/internal/parser"

	"github.com/Masterminds/semver/v3"
)

type Calculator struct {
	config *config.Config
}

func NewCalculator(cfg *config.Config) *Calculator {
	return &Calculator{config: cfg}
}

func (c *Calculator) Calculate(current string, commits []*parser.Commit) (string, config.BumpType, error) {
	ver, err := c.parseVersion(current)
	if err != nil {
		return "", config.BumpNone, fmt.Errorf("invalid current version %s: %w", current, err)
	}

	bump := c.DetermineBump(commits)

	var newVer semver.Version
	switch bump {
	case config.BumpMajor:
		newVer = ver.IncMajor()
	case config.BumpMinor:
		newVer = ver.IncMinor()
	case config.BumpPatch:
		newVer = ver.IncPatch()
	default:
		return current, config.BumpNone, nil
	}

	return c.formatVersion(&newVer), bump, nil
}

func (c *Calculator) DetermineBump(commits []*parser.Commit) config.BumpType {
	bump := config.BumpNone

	for _, commit := range commits {
		if commit.ForceMajor {
			return config.BumpMajor
		}

		commitBump := c.config.GetBumpType(commit.Type)

		bump = maxBump(bump, commitBump)
	}

	return bump
}

func (c *Calculator) parseVersion(versionStr string) (*semver.Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")
	return semver.NewVersion(versionStr)
}

func (c *Calculator) formatVersion(ver *semver.Version) string {
	verStr := ver.String()
	if c.config.Version.Format == "v-prefix" {
		return "v" + verStr
	}
	return verStr
}

func maxBump(a, b config.BumpType) config.BumpType {
	precedence := map[config.BumpType]int{
		config.BumpMajor: 3,
		config.BumpMinor: 2,
		config.BumpPatch: 1,
		config.BumpNone:  0,
	}

	if precedence[a] > precedence[b] {
		return a
	}
	return b
}

func IsValid(versionStr string) bool {
	versionStr = strings.TrimPrefix(versionStr, "v")
	_, err := semver.NewVersion(versionStr)
	return err == nil
}

func Compare(v1, v2 string) (int, error) {
	ver1, err := semver.NewVersion(strings.TrimPrefix(v1, "v"))
	if err != nil {
		return 0, fmt.Errorf("invalid version v1: %w", err)
	}

	ver2, err := semver.NewVersion(strings.TrimPrefix(v2, "v"))
	if err != nil {
		return 0, fmt.Errorf("invalid version v2: %w", err)
	}

	return ver1.Compare(ver2), nil
}
