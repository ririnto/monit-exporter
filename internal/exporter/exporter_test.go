package exporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/commercetools/monit-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
)

// TestNewExporter_NilConfig verifies that providing a nil config returns an error.
func TestNewExporter_NilConfig(t *testing.T) {
	t.Log("Testing NewExporter with nil config")
	exp, err := NewExporter(nil)
	if err == nil {
		t.Errorf("Expected ErrNilConfig, got no error")
	}
	if exp != nil {
		t.Errorf("Expected nil Exporter, got a non-nil instance")
	}
}

// TestNewExporter_ValidConfig verifies that a valid config creates an Exporter successfully.
func TestNewExporter_ValidConfig(t *testing.T) {
	t.Log("Testing NewExporter with a valid config")
	cfg := &config.Config{ListenAddress: "0.0.0.0:9999"}
	exp, err := NewExporter(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exp == nil {
		t.Fatal("Expected a non-nil Exporter, got nil")
	}
}

// TestExporter_Collect_Success uses a mock server returning a valid Monit XML response.
func TestExporter_Collect_Success(t *testing.T) {
	t.Log("Testing Exporter.Collect with a successful Monit response")

	mockXML := `<?xml version="1.0"?>
    <monit>
      <server><version>5.26.0</version></server>
      <platform/>
      <service type="0">
        <name>rootfs</name>
        <status>0</status>
        <monitor>1</monitor>
      </service>
    </monit>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock server received request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, mockXML)
	}))
	defer server.Close()

	cfg := &config.Config{
		MonitScrapeURI: server.URL,
	}
	exp, err := NewExporter(cfg)
	if err != nil {
		t.Fatalf("Failed to create Exporter: %v", err)
	}

	ch := make(chan prometheus.Metric)
	go func() {
		exp.Collect(ch)
		close(ch)
	}()

	for range ch {
	}

	upValue := testutil.ToFloat64(exp.up)
	if upValue != 1 {
		t.Errorf("Expected exporter_up=1, got %f", upValue)
	}
}

// TestExporter_Collect_MonitError uses a mock server that returns an HTTP error.
func TestExporter_Collect_MonitError(t *testing.T) {
	t.Log("Testing Exporter.Collect when Monit returns an error status code")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "some error", http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := &config.Config{MonitScrapeURI: server.URL}
	exp, err := NewExporter(cfg)
	if err != nil {
		t.Fatalf("Failed to create Exporter: %v", err)
	}

	ch := make(chan prometheus.Metric)
	go func() {
		exp.Collect(ch)
		close(ch)
	}()

	upValue := testutil.ToFloat64(exp.up)
	if upValue != 0 {
		t.Errorf("Expected exporter_up=0 on error, got %f", upValue)
	}
}

// Example optional test: verifying logs (only if needed).
func TestExporter_Logs(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	t.Log("Testing Exporter logs at DebugLevel (mock scenario)")
}
