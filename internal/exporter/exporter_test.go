package exporter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/ririnto/monit-exporter/internal/config"
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

	const user, password = "monit-test-user", "monit-metrics-test-password"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotUser, gotPassword, ok := r.BasicAuth(); !ok || gotUser != user || gotPassword != password {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		_, _ = fmt.Fprintln(w, mockXML)
	}))
	defer server.Close()

	cfg := &config.Config{
		MonitScrapeURI: server.URL,
		MonitUser:      user,
		MonitPassword:  password,
	}
	exp, err := NewExporter(cfg)
	if err != nil {
		t.Fatalf("Failed to create Exporter: %v", err)
	}

	if err := testutil.CollectAndCompare(exp, strings.NewReader(`
# HELP monit_exporter_service_check Indicates the status field from Monit.
# TYPE monit_exporter_service_check gauge
monit_exporter_service_check{service_monitor_status="1",service_name="rootfs",service_type="Filesystem"} 0
# HELP monit_exporter_up Indicates whether the Monit endpoint is reachable (1) or not (0).
# TYPE monit_exporter_up gauge
monit_exporter_up 1
`)); err != nil {
		t.Fatal(err)
	}
}

// TestExporter_Collect_MonitError uses a mock server that returns an HTTP error.
func TestExporter_Collect_MonitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "some error", http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := &config.Config{MonitScrapeURI: server.URL}
	exp, err := NewExporter(cfg)
	if err != nil {
		t.Fatalf("Failed to create Exporter: %v", err)
	}

	if err := testutil.CollectAndCompare(exp, strings.NewReader(`
# HELP monit_exporter_up Indicates whether the Monit endpoint is reachable (1) or not (0).
# TYPE monit_exporter_up gauge
monit_exporter_up 0
`)); err != nil {
		t.Fatal(err)
	}
}

// Example optional test: verifying logs (only if needed).
func TestExporter_Logs(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	t.Log("Testing Exporter logs at DebugLevel (mock scenario)")
}
