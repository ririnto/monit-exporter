package exporter

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/ririnto/monit_exporter/internal/config"
)

// TestNewExporter checks if a new Exporter is created without error.
func TestNewExporter(t *testing.T) {
	cfg := &config.Config{}
	exp, err := NewExporter(cfg)
	if err != nil {
		t.Fatalf("failed to create Exporter: %v", err)
	}
	if exp == nil {
		t.Fatal("Exporter is nil")
	}
}

// TestNewExporterNilConfig ensures nil config returns an error.
func TestNewExporterNilConfig(t *testing.T) {
	exp, err := NewExporter(nil)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !errors.Is(err, ErrNilConfig) {
		t.Errorf("expected ErrNilConfig, got %v", err)
	}
	if exp != nil {
		t.Fatal("expected Exporter to be nil")
	}
}

// TestExporterMetrics checks if the Exporter exposes basic metrics.
func TestExporterMetrics(t *testing.T) {
	cfg := &config.Config{}
	exp, _ := NewExporter(cfg)

	metricCount := testutil.CollectAndCount(exp)
	if metricCount == 0 {
		t.Errorf("no metrics collected")
	}
}
