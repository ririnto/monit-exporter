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
	logrus.Debugf("SetLogLevel called with levelStr=%s", levelStr)
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		logrus.Errorf("Failed to parse log level: %v", err)
		return err
	}
	logrus.SetLevel(level)
	logrus.Infof("Log level successfully set to '%s'", level.String())
	return nil
}
