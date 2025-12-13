package logger_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/greysquirr3l/shadowforge/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_DefaultsToCLI(t *testing.T) {
	os.Unsetenv("SHADOWFORGE_LOG_MODE")
	log := logger.NewLogger()
	require.NotNil(t, log)
	_, ok := log.Formatter.(*logrus.TextFormatter)
	assert.True(t, ok, "Default logger should use TextFormatter for CLI")
}

func TestNewCLILogger(t *testing.T) {
	log := logger.NewCLILogger()
	require.NotNil(t, log)
	_, ok := log.Formatter.(*logrus.TextFormatter)
	assert.True(t, ok)
}

func TestNewAPILogger(t *testing.T) {
	log := logger.NewAPILogger()
	require.NotNil(t, log)
	_, ok := log.Formatter.(*logrus.JSONFormatter)
	assert.True(t, ok)
}

func TestSetMode_CLI(t *testing.T) {
	logger.SetMode(logger.ModeCLI)
	assert.Equal(t, logger.ModeCLI, logger.GetMode())
}

func TestSetMode_API(t *testing.T) {
	logger.SetMode(logger.ModeAPI)
	assert.Equal(t, logger.ModeAPI, logger.GetMode())
}

func TestGlobalLoggerFunctions(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.InfoLevel)
	logger.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "test message")
}
