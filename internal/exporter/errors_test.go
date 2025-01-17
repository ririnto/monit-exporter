package exporter

import (
	"errors"
	"testing"
)

// TestErrNilConfig checks if ErrNilConfig is defined as expected.
func TestErrNilConfig(t *testing.T) {
	const wantMsg = "config is nil"

	// 1) Check if the error message matches.
	gotMsg := ErrNilConfig.Error()
	if gotMsg != wantMsg {
		t.Errorf("expected error message: %q, got: %q", wantMsg, gotMsg)
	}

	// 2) Check error type reference (optional).
	//    errors.Is() can be used to verify error chaining or sentinel errors.
	if !errors.Is(ErrNilConfig, ErrNilConfig) {
		t.Errorf("expected errors.Is to confirm ErrNilConfig itself")
	}
}

// TestErrNilConfigUsage demonstrates a typical usage scenario: checking if
// NewExporter returns ErrNilConfig when passed a nil config.
func TestErrNilConfigUsage(t *testing.T) {
	exp, err := NewExporter(nil)
	if exp != nil {
		t.Fatal("expected Exporter to be nil when config is nil")
	}
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
	if !errors.Is(err, ErrNilConfig) {
		t.Errorf("expected err to be ErrNilConfig, got %v", err)
	}
}
