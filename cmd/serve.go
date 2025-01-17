package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ririnto/monit_exporter/internal/config"
	"github.com/ririnto/monit_exporter/internal/exporter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// serveCmd starts the Monit Exporter server.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the Monit Exporter server",
	Long:  "Run the Monit Exporter server that collects Monit status and exposes Prometheus metrics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		if err := config.SetLogLevel(logLevel); err != nil {
			return fmt.Errorf("failed to set log level: %w", err)
		}

		cfg := &config.Config{
			ListenAddress:  listenAddress,
			MetricsPath:    metricsPath,
			IgnoreSSL:      ignoreSSL,
			MonitScrapeURI: monitScrapeURI,
			MonitUser:      monitUser,
			MonitPassword:  monitPassword,
			LogLevel:       logLevel,
		}

		exp, err := exporter.NewExporter(cfg)
		if err != nil {
			return fmt.Errorf("failed to create exporter: %w", err)
		}

		// Register the exporter with Prometheus
		prometheus.MustRegister(exp)

		mux := http.NewServeMux()
		mux.Handle(cfg.MetricsPath, commonLogHandler(promhttp.Handler()))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprintf(
				w,
				`<html>
                <head><title>Monit Exporter</title></head>
                <body>
                <h1>Monit Exporter</h1>
                <p><a href="%s">Metrics</a></p>
                </body>
                </html>`,
				cfg.MetricsPath,
			)
		})

		server := &http.Server{
			Addr:    cfg.ListenAddress,
			Handler: mux,
		}

		// Graceful shutdown setup
		shutdownCh := make(chan os.Signal, 1)
		signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-shutdownCh
			logrus.Info("Received shutdown signal, stopping Monit Exporter...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				logrus.Errorf("Failed to gracefully shutdown: %v", err)
			}
		}()

		logrus.Infof("Starting Monit Exporter on %s", cfg.ListenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to start server: %w", err)
		}
		logrus.Info("Monit Exporter stopped")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

// LoggingResponseWriter wraps a http.ResponseWriter to track status code and size.
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// NewLoggingResponseWriter creates a new LoggingResponseWriter with default status code 200.
func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader sets the status code and calls the underlying ResponseWriter's WriteHeader.
func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Write writes the data and keeps track of the size in bytes.
func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

// commonLogHandler returns an http.Handler that logs requests in Common Log Format.
func commonLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		logrus.Infof("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %.4f",
			r.RemoteAddr,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.RequestURI,
			r.Proto,
			lrw.statusCode,
			lrw.size,
			r.Referer(),
			r.UserAgent(),
			duration.Seconds(),
		)
	})
}
