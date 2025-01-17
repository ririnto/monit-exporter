package config

import (
	"testing"

	"github.com/sirupsen/logrus"
)

// TestSetLogLevelValid tests SetLogLevel with valid log level strings.
func TestSetLogLevelValid(t *testing.T) {
	testCases := []struct {
		levelStr string
		expected logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"warning", logrus.WarnLevel}, // Alias for warn
		{"error", logrus.ErrorLevel},
		{"fatal", logrus.FatalLevel},
		{"panic", logrus.PanicLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.levelStr, func(t *testing.T) {
			err := SetLogLevel(tc.levelStr)
			if err != nil {
				t.Fatalf("SetLogLevel(%q) returned error: %v", tc.levelStr, err)
			}

			got := logrus.GetLevel()
			if got != tc.expected {
				t.Errorf("Expected log level %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestSetLogLevelInvalid tests SetLogLevel with an invalid log level string.
func TestSetLogLevelInvalid(t *testing.T) {
	invalidLevel := "invalid_level"

	err := SetLogLevel(invalidLevel)
	if err == nil {
		t.Fatalf("SetLogLevel(%q) expected to return an error, but got nil", invalidLevel)
	}

	expectedErrMsg := "not a valid logrus Level"
	if !contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain %q, but got %q", expectedErrMsg, err.Error())
	}
}

// contains is a helper function to check if substr is within str.
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > len(substr) && (str[:len(substr)] == substr || contains(str[1:], substr)))
}

// TestSetLogLevelCaseInsensitive tests that SetLogLevel is case-insensitive.
func TestSetLogLevelCaseInsensitive(t *testing.T) {
	testCases := []struct {
		levelStr string
		expected logrus.Level
	}{
		{"DEBUG", logrus.DebugLevel},
		{"Info", logrus.InfoLevel},
		{"WaRn", logrus.WarnLevel},
		{"ErRoR", logrus.ErrorLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.levelStr, func(t *testing.T) {
			err := SetLogLevel(tc.levelStr)
			if err != nil {
				t.Fatalf("SetLogLevel(%q) returned error: %v", tc.levelStr, err)
			}

			got := logrus.GetLevel()
			if got != tc.expected {
				t.Errorf("Expected log level %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestSetLogLevelEmpty tests SetLogLevel with an empty string.
func TestSetLogLevelEmpty(t *testing.T) {
	emptyLevel := ""

	err := SetLogLevel(emptyLevel)
	if err == nil {
		t.Fatalf("SetLogLevel(empty string) expected to return an error, but got nil")
	}

	expectedErrMsg := "not a valid logrus Level"
	if !contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain %q, but got %q", expectedErrMsg, err.Error())
	}
}
