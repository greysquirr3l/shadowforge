package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	assert.NotEmpty(t, version, "Version should not be empty")
	assert.Regexp(t, `^\d+\.\d+\.\d+`, version, "Version should follow semantic versioning")
}

func TestGetBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo()

	assert.NotEmpty(t, buildInfo.Version, "Version should not be empty")
	assert.NotEmpty(t, buildInfo.GoVersion, "Go version should not be empty")
	assert.NotEmpty(t, buildInfo.BuildTime, "Build time should not be empty")

	// GitCommit might be "unknown" in some environments, so just check it's not empty
	assert.NotEmpty(t, buildInfo.GitCommit, "Git commit should not be empty")
}

func TestGetSemanticVersion(t *testing.T) {
	semVer := GetSemanticVersion()
	assert.NotEmpty(t, semVer, "Semantic version should not be empty")

	// Should contain version number
	assert.Contains(t, semVer, GetVersion(), "Semantic version should contain base version")
}
