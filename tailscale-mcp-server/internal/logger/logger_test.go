package logger

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitialize_Default(t *testing.T) {
	err := Initialize(0, "") // Debug level, no file
	require.NoError(t, err)
}

func TestInitialize_WithInvalidLogLevel(t *testing.T) {
	err := Initialize(100, "")
	require.NoError(t, err)
}

func TestInitialize_WithInvalidLogFile(t *testing.T) {
	// Provide a path in non-existent directory to force failure
	err := Initialize(0, "/this/path/likely/does/not/exist/invalid.log")
	require.Error(t, err)
}

func TestHumanReadableConsoleOutput(t *testing.T) {
	// Ensure LOG_FORMAT is unset to get default (console)
	t.Setenv("LOG_FORMAT", "")

	// Create a temporary log file to capture output
	tmpFile, err := os.CreateTemp(t.TempDir(), "testlog-*.log")
	require.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	_ = tmpFile.Close()

	// Re-initialize logger with debug level so all logs flush
	err = Initialize(0, tmpFilePath)
	require.NoError(t, err)

	// Emit a sample log entry
	Info("human readable test", "key", "value")

	// Flush logger to file
	_ = Cleanup()

	// Read back the log file
	data, readErr := os.ReadFile(tmpFilePath)
	require.NoError(t, readErr)

	content := string(data)

	// Console encoding lines should contain tabs and not start with '{'
	require.NotEmpty(t, content)

	lines := strings.Split(strings.TrimSpace(content), "\n")
	// Check first non-empty line
	var firstLine string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			firstLine = l
			break
		}
	}
	require.NotEmpty(t, firstLine)

	if firstLine[0] == '{' {
		t.Fatalf("expected console (human-readable) encoding, got JSON: %s", firstLine)
	}

	// Ensure key/value appear in line for readability
	require.Contains(t, firstLine, "human readable test")
	require.Contains(t, firstLine, "key")
	require.Contains(t, firstLine, "value")
}
