package config

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSetLogLevel_Success(t *testing.T) {
	err := SetLogLevel("debug")
	if err != nil {
		t.Fatalf("Expected no error for valid level 'debug', got %v", err)
	}
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Errorf("Expected log level=DebugLevel, got %s", logrus.GetLevel())
	}
}

func TestSetLogLevel_Invalid(t *testing.T) {
	err := SetLogLevel("notalevel")
	if err == nil {
		t.Fatal("Expected an error for invalid log level 'notalevel', got nil")
	}
}
