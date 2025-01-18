package exporter

import (
	"errors"
	"slices"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ririnto/monit-exporter/internal/config"
	"github.com/ririnto/monit-exporter/internal/monit"
	"github.com/sirupsen/logrus"
)

const (
	namespace = "monit"
)

var (
	// ErrNilConfig is returned when a nil config is provided to NewExporter.
	ErrNilConfig = errors.New("config is nil")
)

// serviceTypes maps Monit service type integers to descriptive strings.
var serviceTypes = map[int]string{
	0: "Filesystem",
	1: "Directory",
	2: "File",
	3: "Process",
	4: "Remote host",
	5: "System",
	6: "Fifo",
	7: "Program",
	8: "Network",
}

// Exporter collects Monit metrics and exposes them to Prometheus.
type Exporter struct {
	cfg   *config.Config
	mutex sync.Mutex

	up     prometheus.Gauge
	status *prometheus.GaugeVec

	blockUsage   *prometheus.GaugeVec
	blockTotal   *prometheus.GaugeVec
	blockPercent *prometheus.GaugeVec

	inodeUsage   *prometheus.GaugeVec
	inodeTotal   *prometheus.GaugeVec
	inodePercent *prometheus.GaugeVec

	portResponseTime *prometheus.GaugeVec

	systemLoadAvg01 *prometheus.GaugeVec
	systemLoadAvg05 *prometheus.GaugeVec
	systemLoadAvg15 *prometheus.GaugeVec

	systemCPUUser   *prometheus.GaugeVec
	systemCPUSystem *prometheus.GaugeVec
	systemCPUWait   *prometheus.GaugeVec

	systemMemPercent    *prometheus.GaugeVec
	systemMemKilobytes  *prometheus.GaugeVec
	systemSwapPercent   *prometheus.GaugeVec
	systemSwapKilobytes *prometheus.GaugeVec
}

// NewExporter creates a new Exporter using the given Config.
func NewExporter(cfg *config.Config) (*Exporter, error) {
	if cfg == nil {
		logrus.Error("NewExporter: config is nil")
		return nil, ErrNilConfig
	}

	logrus.Debugf("NewExporter: creating exporter with ListenAddress=%s, MonitScrapeURI=%s",
		cfg.ListenAddress, cfg.MonitScrapeURI)

	labelNames := []string{"service_name", "service_type", "service_monitor_status"}

	return &Exporter{
		cfg: cfg,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "exporter_up",
			Help:      "Indicates whether the Monit endpoint is reachable (1) or not (0).",
		}),
		status: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "exporter_service_check",
				Help:      "Indicates the status field from Monit.",
			},
			labelNames,
		),

		blockUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_block_usage_bytes",
				Help:      "Block usage for filesystem-based services.",
			},
			labelNames,
		),
		blockTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_block_total_bytes",
				Help:      "Block total capacity for filesystem-based services.",
			},
			labelNames,
		),
		blockPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_block_usage_percent",
				Help:      "Block usage percentage for filesystem-based services.",
			},
			labelNames,
		),

		inodeUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_inode_usage",
				Help:      "Inode usage for filesystem-based services.",
			},
			labelNames,
		),
		inodeTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_inode_total",
				Help:      "Total number of inodes for filesystem-based services.",
			},
			labelNames,
		),
		inodePercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_inode_usage_percent",
				Help:      "Inode usage percentage for filesystem-based services.",
			},
			labelNames,
		),

		portResponseTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_port_response_seconds",
				Help:      "Response time in seconds for port-based checks.",
			},
			labelNames,
		),

		systemLoadAvg01: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_loadavg_01",
				Help:      "1-minute load average for system-based services.",
			},
			labelNames,
		),
		systemLoadAvg05: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_loadavg_05",
				Help:      "5-minute load average for system-based services.",
			},
			labelNames,
		),
		systemLoadAvg15: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_loadavg_15",
				Help:      "15-minute load average for system-based services.",
			},
			labelNames,
		),

		systemCPUUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_cpu_user_percent",
				Help:      "CPU usage in user space (percent).",
			},
			labelNames,
		),
		systemCPUSystem: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_cpu_system_percent",
				Help:      "CPU usage in kernel space (percent).",
			},
			labelNames,
		),
		systemCPUWait: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_cpu_wait_percent",
				Help:      "CPU usage waiting for I/O (percent).",
			},
			labelNames,
		),

		systemMemPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_memory_usage_percent",
				Help:      "Memory usage percentage for system-based services.",
			},
			labelNames,
		),
		systemMemKilobytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_memory_usage_kilobytes",
				Help:      "Memory usage in kilobytes for system-based services.",
			},
			labelNames,
		),
		systemSwapPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_swap_usage_percent",
				Help:      "Swap usage percentage for system-based services.",
			},
			labelNames,
		),
		systemSwapKilobytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "service_system_swap_usage_kilobytes",
				Help:      "Swap usage in kilobytes for system-based services.",
			},
			labelNames,
		),
	}, nil
}

// Describe sends the descriptors of each metric to the provided channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.status.Describe(ch)

	e.blockUsage.Describe(ch)
	e.blockTotal.Describe(ch)
	e.blockPercent.Describe(ch)

	e.inodeUsage.Describe(ch)
	e.inodeTotal.Describe(ch)
	e.inodePercent.Describe(ch)

	e.portResponseTime.Describe(ch)

	e.systemLoadAvg01.Describe(ch)
	e.systemLoadAvg05.Describe(ch)
	e.systemLoadAvg15.Describe(ch)

	e.systemCPUUser.Describe(ch)
	e.systemCPUSystem.Describe(ch)
	e.systemCPUWait.Describe(ch)

	e.systemMemPercent.Describe(ch)
	e.systemMemKilobytes.Describe(ch)
	e.systemSwapPercent.Describe(ch)
	e.systemSwapKilobytes.Describe(ch)

	logrus.Debug("Exporter.Describe: described all metrics to the channel")
}

// Collect is called by the Prometheus registry to gather metrics.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	logrus.Debug("Exporter.Collect: resetting metrics before scrape")

	e.status.Reset()
	e.blockUsage.Reset()
	e.blockTotal.Reset()
	e.blockPercent.Reset()
	e.inodeUsage.Reset()
	e.inodeTotal.Reset()
	e.inodePercent.Reset()
	e.portResponseTime.Reset()
	e.systemLoadAvg01.Reset()
	e.systemLoadAvg05.Reset()
	e.systemLoadAvg15.Reset()
	e.systemCPUUser.Reset()
	e.systemCPUSystem.Reset()
	e.systemCPUWait.Reset()
	e.systemMemPercent.Reset()
	e.systemMemKilobytes.Reset()
	e.systemSwapPercent.Reset()
	e.systemSwapKilobytes.Reset()

	err := e.scrape()
	if err != nil {
		logrus.Errorf("Exporter.Collect: scrape error: %v", err)
	}

	e.up.Collect(ch)
	e.status.Collect(ch)
	e.blockUsage.Collect(ch)
	e.blockTotal.Collect(ch)
	e.blockPercent.Collect(ch)
	e.inodeUsage.Collect(ch)
	e.inodeTotal.Collect(ch)
	e.inodePercent.Collect(ch)
	e.portResponseTime.Collect(ch)
	e.systemLoadAvg01.Collect(ch)
	e.systemLoadAvg05.Collect(ch)
	e.systemLoadAvg15.Collect(ch)
	e.systemCPUUser.Collect(ch)
	e.systemCPUSystem.Collect(ch)
	e.systemCPUWait.Collect(ch)
	e.systemMemPercent.Collect(ch)
	e.systemMemKilobytes.Collect(ch)
	e.systemSwapPercent.Collect(ch)
	e.systemSwapKilobytes.Collect(ch)

	logrus.Debug("Exporter.Collect: metrics collected and sent to the channel")
}

// scrape fetches Monit status and updates the metrics.
func (e *Exporter) scrape() error {
	logrus.Debug("Exporter.scrape: fetching Monit status")
	data, err := monit.FetchMonitStatus(e.cfg)
	if err != nil {
		logrus.Warnf("Exporter.scrape: failed to fetch Monit status: %v", err)
		e.up.Set(0)
		e.status.Reset()
		return err
	}
	logrus.Debugf("Exporter.scrape: successfully fetched Monit status (%d bytes)", len(data))

	parsed, err := monit.ParseMonitStatus(data)
	if err != nil {
		logrus.Warnf("Exporter.scrape: failed to parse Monit status: %v", err)
		e.up.Set(0)
		e.status.Reset()
		return err
	}
	logrus.Debug("Exporter.scrape: successfully parsed Monit status")

	e.up.Set(1)
	logrus.Debug("Exporter.scrape: set exporter_up to 1 (Monit is reachable)")

	for service := range slices.Values(parsed.Services) {
		serviceType, ok := serviceTypes[service.Type]
		if !ok {
			serviceType = "unknown"
			logrus.Warnf("Exporter.scrape: unknown service service_type=%d, serviceNameservice_name=%s", service.Type, service.Name)
		}
		serviceMonitorStatus := strconv.Itoa(service.Monitor)

		e.status.With(prometheus.Labels{
			"service_name":           service.Name,
			"service_type":           serviceType,
			"service_monitor_status": serviceMonitorStatus,
		}).Set(float64(service.Status))

		logrus.Debugf(
			"Exporter.scrape: service_name=%s, service_type=%s, service_monitor_status=%d, service_status=%d",
			service.Name,
			serviceType,
			service.Monitor,
			service.Status,
		)

		e.collectServiceMetrics(service, serviceType, serviceMonitorStatus)
	}
	return nil
}

// collectServiceMetrics updates detailed metrics for a single Monit service.
func (e *Exporter) collectServiceMetrics(service monit.Service, serviceType, serviceMonitorStatus string) {
	labels := prometheus.Labels{
		"service_name":           service.Name,
		"service_type":           serviceType,
		"service_monitor_status": serviceMonitorStatus,
	}

	if service.Block != nil {
		e.blockUsage.With(labels).Set(service.Block.Usage)
		e.blockTotal.With(labels).Set(service.Block.Total)
		e.blockPercent.With(labels).Set(service.Block.Percent)
	}

	if service.Inode != nil {
		e.inodeUsage.With(labels).Set(float64(service.Inode.Usage))
		e.inodeTotal.With(labels).Set(float64(service.Inode.Total))
		e.inodePercent.With(labels).Set(service.Inode.Percent)
	}

	if service.Port != nil {
		e.portResponseTime.With(labels).Set(service.Port.Responsetime)
	}

	if service.System != nil {
		e.systemLoadAvg01.With(labels).Set(service.System.Load.Avg01)
		e.systemLoadAvg05.With(labels).Set(service.System.Load.Avg05)
		e.systemLoadAvg15.With(labels).Set(service.System.Load.Avg15)

		e.systemCPUUser.With(labels).Set(service.System.CPU.User)
		e.systemCPUSystem.With(labels).Set(service.System.CPU.System)
		e.systemCPUWait.With(labels).Set(service.System.CPU.Wait)

		e.systemMemPercent.With(labels).Set(service.System.Memory.Percent)
		e.systemMemKilobytes.With(labels).Set(float64(service.System.Memory.Kilobyte))
		e.systemSwapPercent.With(labels).Set(service.System.Swap.Percent)
		e.systemSwapKilobytes.With(labels).Set(float64(service.System.Swap.Kilobyte))
	}
}
