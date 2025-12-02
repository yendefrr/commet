package version

import (
	"testing"

	"github.com/yendefrr/commet/internal/config"
	"github.com/yendefrr/commet/internal/parser"
)

func TestCalculate(t *testing.T) {
	cfg := &config.Config{
		Version: config.VersionConfig{
			Format: "semver",
		},
		BumpRules: map[string]config.BumpType{
			"Fix":      config.BumpPatch,
			"Feature":  config.BumpMinor,
			"Refactor": config.BumpPatch,
			"Breaking": config.BumpMajor,
		},
	}

	calc := NewCalculator(cfg)

	tests := []struct {
		name            string
		currentVersion  string
		commits         []*parser.Commit
		expectedVersion string
		expectedBump    config.BumpType
	}{
		{
			name:           "single patch",
			currentVersion: "1.2.3",
			commits: []*parser.Commit{
				{Type: "Fix", Description: "fix bug"},
			},
			expectedVersion: "1.2.4",
			expectedBump:    config.BumpPatch,
		},
		{
			name:           "single minor",
			currentVersion: "1.2.3",
			commits: []*parser.Commit{
				{Type: "Feature", Description: "new feature"},
			},
			expectedVersion: "1.3.0",
			expectedBump:    config.BumpMinor,
		},
		{
			name:           "single major",
			currentVersion: "1.2.3",
			commits: []*parser.Commit{
				{Type: "Breaking", Description: "breaking change"},
			},
			expectedVersion: "2.0.0",
			expectedBump:    config.BumpMajor,
		},
		{
			name:           "force major with exclamation",
			currentVersion: "1.2.3",
			commits: []*parser.Commit{
				{Type: "Fix", Description: "critical fix", ForceMajor: true},
			},
			expectedVersion: "2.0.0",
			expectedBump:    config.BumpMajor,
		},
		{
			name:           "multiple commits - highest wins",
			currentVersion: "1.2.3",
			commits: []*parser.Commit{
				{Type: "Fix", Description: "fix bug"},
				{Type: "Feature", Description: "new feature"},
				{Type: "Refactor", Description: "refactor code"},
			},
			expectedVersion: "1.3.0",
			expectedBump:    config.BumpMinor,
		},
		{
			name:           "major overrides all",
			currentVersion: "1.2.3",
			commits: []*parser.Commit{
				{Type: "Fix", Description: "fix bug"},
				{Type: "Feature", Description: "new feature"},
				{Type: "Breaking", Description: "breaking change"},
			},
			expectedVersion: "2.0.0",
			expectedBump:    config.BumpMajor,
		},
		{
			name:           "no bump",
			currentVersion: "1.2.3",
			commits:        []*parser.Commit{},
			expectedVersion: "1.2.3",
			expectedBump:    config.BumpNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, bump, err := calc.Calculate(tt.currentVersion, tt.commits)
			if err != nil {
				t.Errorf("Calculate() error = %v", err)
				return
			}

			if version != tt.expectedVersion {
				t.Errorf("Calculate() version = %v, want %v", version, tt.expectedVersion)
			}

			if bump != tt.expectedBump {
				t.Errorf("Calculate() bump = %v, want %v", bump, tt.expectedBump)
			}
		})
	}
}

func TestCalculateWithVPrefix(t *testing.T) {
	cfg := &config.Config{
		Version: config.VersionConfig{
			Format: "v-prefix",
		},
		BumpRules: map[string]config.BumpType{
			"Feature": config.BumpMinor,
		},
	}

	calc := NewCalculator(cfg)

	version, bump, err := calc.Calculate("v1.2.3", []*parser.Commit{
		{Type: "Feature", Description: "new feature"},
	})

	if err != nil {
		t.Errorf("Calculate() error = %v", err)
		return
	}

	if version != "v1.3.0" {
		t.Errorf("Calculate() version = %v, want v1.3.0", version)
	}

	if bump != config.BumpMinor {
		t.Errorf("Calculate() bump = %v, want %v", bump, config.BumpMinor)
	}
}

func TestDetermineBump(t *testing.T) {
	cfg := &config.Config{
		BumpRules: map[string]config.BumpType{
			"Fix":      config.BumpPatch,
			"Feature":  config.BumpMinor,
			"Breaking": config.BumpMajor,
			"Docs":     config.BumpNone,
		},
	}

	calc := NewCalculator(cfg)

	tests := []struct {
		name         string
		commits      []*parser.Commit
		expectedBump config.BumpType
	}{
		{
			name: "patch only",
			commits: []*parser.Commit{
				{Type: "Fix"},
			},
			expectedBump: config.BumpPatch,
		},
		{
			name: "minor only",
			commits: []*parser.Commit{
				{Type: "Feature"},
			},
			expectedBump: config.BumpMinor,
		},
		{
			name: "major only",
			commits: []*parser.Commit{
				{Type: "Breaking"},
			},
			expectedBump: config.BumpMajor,
		},
		{
			name: "force major",
			commits: []*parser.Commit{
				{Type: "Fix", ForceMajor: true},
			},
			expectedBump: config.BumpMajor,
		},
		{
			name: "mixed - minor wins over patch",
			commits: []*parser.Commit{
				{Type: "Fix"},
				{Type: "Feature"},
			},
			expectedBump: config.BumpMinor,
		},
		{
			name: "mixed - major wins over all",
			commits: []*parser.Commit{
				{Type: "Fix"},
				{Type: "Feature"},
				{Type: "Breaking"},
			},
			expectedBump: config.BumpMajor,
		},
		{
			name: "none only",
			commits: []*parser.Commit{
				{Type: "Docs"},
			},
			expectedBump: config.BumpNone,
		},
		{
			name:         "no commits",
			commits:      []*parser.Commit{},
			expectedBump: config.BumpNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bump := calc.DetermineBump(tt.commits)
			if bump != tt.expectedBump {
				t.Errorf("DetermineBump() = %v, want %v", bump, tt.expectedBump)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.2.3", true},
		{"v1.2.3", true},
		{"0.0.1", true},
		{"10.20.30", true},
		{"1.2", true},    // semver library accepts this as 1.2.0
		{"1", true},      // semver library accepts this as 1.0.0
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := IsValid(tt.version); got != tt.valid {
				t.Errorf("IsValid(%v) = %v, want %v", tt.version, got, tt.valid)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3", "1.2.3", 0},
		{"1.2.3", "1.2.4", -1},
		{"1.2.4", "1.2.3", 1},
		{"1.3.0", "1.2.9", 1},
		{"2.0.0", "1.9.9", 1},
		{"v1.2.3", "v1.2.3", 0},
		{"v1.2.3", "1.2.3", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			result, err := Compare(tt.v1, tt.v2)
			if err != nil {
				t.Errorf("Compare() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Compare(%v, %v) = %v, want %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}
