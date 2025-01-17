package exporter

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ririnto/monit_exporter/internal/config"
	"github.com/ririnto/monit_exporter/internal/monit"
	"github.com/sirupsen/logrus"
)

const (
	namespace = "monit"
)

// serviceTypes maps Monit service type integers to descriptive strings.
var serviceTypes = map[int]string{
	0: "filesystem",
	1: "directory",
	2: "file",
	3: "program_with_pidfile",
	4: "remote_host",
	5: "system",
	6: "fifo",
	7: "program_with_path",
	8: "network",
}

// Exporter collects Monit metrics and exposes them to Prometheus.
type Exporter struct {
	cfg    *config.Config
	mutex  sync.Mutex
	up     prometheus.Gauge
	status *prometheus.GaugeVec
}

// NewExporter creates a new Exporter using the given Config.
func NewExporter(cfg *config.Config) (*Exporter, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

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
				Help:      "Monit service check info. The gauge value is the 'status' field from Monit.",
			},
			[]string{"check_name", "type", "monitored"},
		),
	}, nil
}

// Describe sends the descriptors of each metric over to the provided channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.up.Describe(ch)
	e.status.Describe(ch)
}

// Collect is called by the Prometheus registry to gather metrics.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.status.Reset()

	if err := e.scrape(); err != nil {
		logrus.Errorf("Error scraping Monit: %v", err)
	}

	e.up.Collect(ch)
	e.status.Collect(ch)
}

// scrape fetches and parses the Monit status and updates the metrics.
func (e *Exporter) scrape() error {
	data, err := monit.FetchMonitStatus(e.cfg)
	if err != nil {
		e.up.Set(0)
		e.status.Reset()
		return err
	}

	parsed, err := monit.ParseMonitStatus(data)
	if err != nil {
		e.up.Set(0)
		e.status.Reset()
		return err
	}

	e.up.Set(1)
	for _, svc := range parsed.Services {
		typ, ok := serviceTypes[svc.Type]
		if !ok {
			typ = "unknown"
		}
		e.status.With(prometheus.Labels{
			"check_name": svc.Name,
			"type":       typ,
			"monitored":  svc.Monitored,
		}).Set(float64(svc.Status))
	}

	return nil
}
