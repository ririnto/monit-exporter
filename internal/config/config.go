package config

import (
	"github.com/sirupsen/logrus"
)

// Config holds the configuration values needed by the Monit Exporter.
type Config struct {
	ListenAddress  string
	MetricsPath    string
	IgnoreSSL      bool
	MonitScrapeURI string
	MonitUser      string
	MonitPassword  string
	LogLevel       string
}

// SetLogLevel sets the global log level of logrus based on the given string.
func SetLogLevel(levelStr string) error {
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	return nil
}
