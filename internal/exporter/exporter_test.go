package exporter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func TestExporter_ScrapeProcessPID(t *testing.T) {
	var services atomic.Pointer[string]
	monitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `<monit><server><version>5.26.0</version></server>%s</monit>`, *services.Load())
	}))
	defer monitServer.Close()

	exp, err := NewExporter(&config.Config{MonitScrapeURI: monitServer.URL})
	if err != nil {
		t.Fatal(err)
	}
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(exp)
	metricsServer := httptest.NewServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	defer metricsServer.Close()

	for _, scrape := range []struct {
		name     string
		services string
		pid      int
	}{
		{
			name: "reports only valid process PIDs",
			services: `
<service type="3"><name>worker</name><monitor>1</monitor><pid> 42 </pid></service>
<service type="3"><name>empty</name><pid/></service>
<service type="3"><name>zero</name><pid>0</pid></service>
<service type="3"><name>negative</name><pid>-1</pid></service>
<service type="3"><name>nonnumeric</name><pid>unknown</pid></service>
<service type="3"><name>out-of-range</name><pid>2147483648</pid></service>
<service type="0"><name>rootfs</name><pid>73</pid></service>`,
			pid: 42,
		},
		{
			name:     "updates PID in the same series when monitor status changes",
			services: `<service type="3"><name>worker</name><monitor>2</monitor><pid>57</pid></service>`,
			pid:      57,
		},
		{
			name:     "removes PID when Monit omits it",
			services: `<service type="3"><name>worker</name><monitor>1</monitor></service>`,
		},
		{
			name:     "reports PID again when available",
			services: `<service type="3"><name>worker</name><monitor>1</monitor><pid>81</pid></service>`,
			pid:      81,
		},
		{name: "removes PID when the service disappears"},
	} {
		t.Run(scrape.name, func(t *testing.T) {
			services.Store(&scrape.services)
			expected := `
# HELP monit_exporter_up Indicates whether the Monit endpoint is reachable (1) or not (0).
# TYPE monit_exporter_up gauge
monit_exporter_up 1
`
			if scrape.pid > 0 {
				expected += fmt.Sprintf(`
# HELP monit_service_process_pid Process ID reported by Monit.
# TYPE monit_service_process_pid gauge
monit_service_process_pid{service_name="worker"} %d
`, scrape.pid)
			}
			if err := testutil.ScrapeAndCompare(metricsServer.URL+"/metrics", strings.NewReader(expected),
				"monit_exporter_up", "monit_service_process_pid"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExporter_ScrapeProcessResources(t *testing.T) {
	const process = `<service type="3">
  <name>worker</name><status>0</status><monitor>1</monitor><pid>42</pid>
  <memory><percent>1.0</percent><percenttotal>2.0</percenttotal><kilobyte>1536</kilobyte><kilobytetotal>4096</kilobytetotal></memory>
  <cpu><percent>2.5</percent><percenttotal>9.5</percenttotal></cpu>
  <write><bytes><count>256</count><total>4294967296</total></bytes></write>
</service>`
	const filesystem = `<service type="0"><name>rootfs</name><status>0</status><monitor>1</monitor>
  <write><bytes><count>4096</count><total>8192</total></bytes></write>
</service>`
	const up = `
# HELP monit_exporter_up Indicates whether the Monit endpoint is reachable (1) or not (0).
# TYPE monit_exporter_up gauge
monit_exporter_up 1
`
	const resources = `
# HELP monit_service_cpu_perc CPU usage percentage for a Monit process check.
# TYPE monit_service_cpu_perc gauge
monit_service_cpu_perc{check_name="worker",type="percentage"} 2.5
monit_service_cpu_perc{check_name="worker",type="percentage_total"} 9.5
# HELP monit_service_mem_bytes Memory usage in bytes for a Monit process check.
# TYPE monit_service_mem_bytes gauge
monit_service_mem_bytes{check_name="worker",type="process"} 1572864
monit_service_mem_bytes{check_name="worker",type="process_and_children"} 4194304
# HELP monit_service_write_bytes_per_second Disk write rate in bytes per second for a Monit process check.
# TYPE monit_service_write_bytes_per_second gauge
monit_service_write_bytes_per_second{check_name="worker"} 256
# HELP monit_service_write_cumulative_bytes Cumulative disk bytes written as reported by Monit for a process check.
# TYPE monit_service_write_cumulative_bytes gauge
monit_service_write_cumulative_bytes{check_name="worker"} 4294967296
`
	var services atomic.Pointer[string]
	monitServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if body := services.Load(); body != nil {
			_, _ = fmt.Fprintf(w, `<monit version="5.26.0"><server/>%s</monit>`, *body)
		} else {
			http.Error(w, "Monit unavailable", http.StatusServiceUnavailable)
		}
	}))
	defer monitServer.Close()

	exp, err := NewExporter(&config.Config{MonitScrapeURI: monitServer.URL})
	if err != nil {
		t.Fatal(err)
	}
	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(exp)
	metricsServer := httptest.NewServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	defer metricsServer.Close()

	for _, scrape := range []struct {
		name     string
		services string
		expected string
	}{
		{
			name:     "reports process resources in Monit units",
			services: process + filesystem,
			expected: up + resources,
		},
		{
			name: "keeps zero and drops unavailable fields",
			services: `<service type="3">
  <name>worker</name><monitor>1</monitor><pid>42</pid>
  <cpu><percent>0.0</percent></cpu>
  <memory><kilobyte/></memory>
  <write><operations><count>3</count><total>34</total></operations></write>
</service>` + filesystem,
			expected: up + `
# HELP monit_service_cpu_perc CPU usage percentage for a Monit process check.
# TYPE monit_service_cpu_perc gauge
monit_service_cpu_perc{check_name="worker",type="percentage"} 0
`,
		},
		{
			name: "reports the write rate when Monit omits cumulative bytes",
			services: `<service type="3"><name>worker</name><monitor>1</monitor>
  <write><bytes><count>0</count></bytes></write>
</service>`,
			expected: up + `
# HELP monit_service_write_bytes_per_second Disk write rate in bytes per second for a Monit process check.
# TYPE monit_service_write_bytes_per_second gauge
monit_service_write_bytes_per_second{check_name="worker"} 0
`,
		},
		{
			name:     "drops resources when the process service disappears",
			services: filesystem,
			expected: up,
		},
		{
			name:     "reports resources again when Monit supplies them",
			services: process,
			expected: up + resources,
		},
		{
			name: "drops resources when the Monit scrape fails",
			expected: `
# HELP monit_exporter_up Indicates whether the Monit endpoint is reachable (1) or not (0).
# TYPE monit_exporter_up gauge
monit_exporter_up 0
`,
		},
	} {
		t.Run(scrape.name, func(t *testing.T) {
			if scrape.services == "" {
				services.Store(nil)
			} else {
				services.Store(&scrape.services)
			}
			if err := testutil.ScrapeAndCompare(metricsServer.URL+"/metrics", strings.NewReader(scrape.expected),
				"monit_exporter_up", "monit_service_cpu_perc", "monit_service_mem_bytes",
				"monit_service_write_bytes_per_second", "monit_service_write_cumulative_bytes"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestExporter_Collect_MonitError uses a mock server that returns an HTTP error.
func TestExporter_Collect_MonitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
