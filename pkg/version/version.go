// Package version provides comprehensive version information for Shadowforge.
// It embeds the VERSION file and provides utilities for accessing version data
// throughout the application with build-time injection support.
package version

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//go:embed VERSION
var versionFile string

const unknownValue = "unknown"

// BuildInfo contains comprehensive version and build information
type BuildInfo struct {
	Version    string `json:"version"`
	GitCommit  string `json:"git_commit"`
	GitTag     string `json:"git_tag"`
	GitBranch  string `json:"git_branch"`
	GitDirty   bool   `json:"git_dirty"`
	BuildTime  string `json:"build_time"`
	BuildUser  string `json:"build_user"`
	BuildHost  string `json:"build_host"`
	GoVersion  string `json:"go_version"`
	GoArch     string `json:"go_arch"`
	GoOS       string `json:"go_os"`
	Compiler   string `json:"compiler"`
	CGOEnabled bool   `json:"cgo_enabled"`
}

// These variables can be set at build time using -ldflags
// Example: go build -ldflags "-X github.com/yourorg/shadowforge/pkg/version.GitCommit=$(git rev-parse HEAD)"
var (
	GitCommit  = unknownValue
	GitTag     = unknownValue
	GitBranch  = unknownValue
	GitDirty   = unknownValue
	BuildTime  = unknownValue
	BuildUser  = unknownValue
	BuildHost  = unknownValue
	CGOEnabled = unknownValue
)

// GetVersion reads the version from the embedded VERSION file
func GetVersion() string {
	return strings.TrimSpace(versionFile)
}

// GetGitCommit gets the current git commit hash
func GetGitCommit() string {
	if GitCommit != unknownValue {
		return GitCommit // Use build-time injected value if available
	}

	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return unknownValue
	}
	return strings.TrimSpace(string(output))
}

// GetGitTag gets the current git tag (if on a tagged commit)
func GetGitTag() string {
	if GitTag != unknownValue {
		return GitTag // Use build-time injected value if available
	}

	cmd := exec.CommandContext(context.Background(), "git", "describe", "--tags", "--exact-match", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "" // Not on a tagged commit
	}
	return strings.TrimSpace(string(output))
}

// GetGitBranch gets the current git branch
func GetGitBranch() string {
	if GitBranch != unknownValue {
		return GitBranch // Use build-time injected value if available
	}

	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return unknownValue
	}
	return strings.TrimSpace(string(output))
}

// IsGitDirty checks if the git working directory is dirty (has uncommitted changes)
func IsGitDirty() bool {
	if GitDirty != unknownValue {
		// Parse build-time injected value
		dirty, err := strconv.ParseBool(GitDirty)
		if err == nil {
			return dirty
		}
	}

	cmd := exec.CommandContext(context.Background(), "git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false // Assume clean if we can't determine
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// GetBuildTime returns the build time
func GetBuildTime() string {
	if BuildTime != unknownValue {
		return BuildTime // Use build-time injected value if available
	}
	return time.Now().UTC().Format(time.RFC3339)
}

// GetBuildUser returns the user who built the binary
func GetBuildUser() string {
	if BuildUser != unknownValue {
		return BuildUser // Use build-time injected value if available
	}
	return unknownValue
}

// GetBuildHost returns the host where the binary was built
func GetBuildHost() string {
	if BuildHost != unknownValue {
		return BuildHost // Use build-time injected value if available
	}
	return unknownValue
}

// GetCGOEnabled returns whether CGO was enabled during build
func GetCGOEnabled() bool {
	if CGOEnabled != unknownValue {
		enabled, err := strconv.ParseBool(CGOEnabled)
		if err == nil {
			return enabled
		}
	}
	// Default to false if unknown
	return false
}

// GetBuildInfo returns complete build information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:    GetVersion(),
		GitCommit:  GetGitCommit(),
		GitTag:     GetGitTag(),
		GitBranch:  GetGitBranch(),
		GitDirty:   IsGitDirty(),
		BuildTime:  GetBuildTime(),
		BuildUser:  GetBuildUser(),
		BuildHost:  GetBuildHost(),
		GoVersion:  runtime.Version(),
		GoArch:     runtime.GOARCH,
		GoOS:       runtime.GOOS,
		Compiler:   runtime.Compiler,
		CGOEnabled: GetCGOEnabled(),
	}
}

// GetSemanticVersion returns a semantic version with git info for development builds
func GetSemanticVersion() string {
	version := GetVersion()

	// Check if we're on a tag
	cmd := exec.CommandContext(context.Background(), "git", "describe", "--tags", "--exact-match", "HEAD")
	if _, err := cmd.Output(); err == nil {
		return version // Clean version on tag
	}

	// Add commit info for non-tag builds
	gitCommit := GetGitCommit()
	if gitCommit != unknownValue {
		return fmt.Sprintf("%s-dev+%s", version, gitCommit)
	}

	return fmt.Sprintf("%s-dev", version)
}

// PrintVersion prints version information to stdout
func PrintVersion() {
	info := GetBuildInfo()
	fmt.Printf("Version:    %s\n", info.Version)
	fmt.Printf("Git Commit: %s\n", info.GitCommit)
	fmt.Printf("Build Time: %s\n", info.BuildTime)
	fmt.Printf("Go Version: %s\n", info.GoVersion)
}
