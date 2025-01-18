package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "embed"

	"github.com/commercetools/monit-exporter/internal/config"
	"github.com/commercetools/monit-exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//go:embed static/favicon.ico
var embeddedFavicon []byte

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the Monit Exporter server",
	Long:  "Run the Monit Exporter server that collects Monit status and exposes Prometheus metrics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.Debug("serveCmd invoked: starting Monit Exporter server")

		if err := config.SetLogLevel(logLevel); err != nil {
			logrus.Errorf("Failed to set log level: %v", err)
			return fmt.Errorf("failed to set log level: %w", err)
		}
		logrus.Debugf("Log level set to '%s'", logLevel)

		cfg := &config.Config{
			ListenAddress:  listenAddress,
			MetricsPath:    metricsPath,
			IgnoreSSL:      ignoreSSL,
			MonitScrapeURI: monitScrapeURI,
			MonitUser:      monitUser,
			MonitPassword:  monitPassword,
			LogLevel:       logLevel,
		}
		logrus.Debugf("Server configuration loaded: %+v", cfg)

		exp, err := exporter.NewExporter(cfg)
		if err != nil {
			logrus.Errorf("Failed to create exporter: %v", err)
			return fmt.Errorf("failed to create exporter: %w", err)
		}
		logrus.Debug("Registering exporter to Prometheus")
		prometheus.MustRegister(exp)

		mux := http.NewServeMux()
		mux.Handle(cfg.MetricsPath, promhttp.Handler())
		mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			if 0 < len(embeddedFavicon) {
				w.Header().Set("Content-Type", "image/x-icon")
				_, _ = w.Write(embeddedFavicon)
			} else {
				http.ServeFile(w, r, filepath.Join("static", "favicon.ico"))
			}
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			logrus.Debugf("Root path request received from %s", r.RemoteAddr)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprintf(
				w,
				`<html><head><title>Monit Exporter</title></head><body><h1>Monit Exporter</h1><p><a href="%s">Metrics</a></p></body></html>`,
				cfg.MetricsPath,
			)
		})

		server := &http.Server{Addr: cfg.ListenAddress, Handler: commonLogHandler(mux)}
		shutdownCh := make(chan os.Signal, 1)
		signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			sig := <-shutdownCh
			logrus.Infof("Received shutdown signal: %v. Attempting to stop Monit Exporter gracefully...", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				logrus.Errorf("Graceful shutdown failed: %v", err)
			} else {
				logrus.Info("Server shut down gracefully")
			}
		}()

		logrus.Infof("Starting Monit Exporter on %s", cfg.ListenAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("Failed to start server: %v", err)
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
		logrus.Infof("[commonLogHandler] %s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %.4f",
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
		logrus.Debugf("[commonLogHandler] Request processed: Method=%s, URI=%s, Duration=%.4fs", r.Method, r.RequestURI, duration.Seconds())
	})
}
