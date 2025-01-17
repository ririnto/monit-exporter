package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listenAddress  string
	metricsPath    string
	ignoreSSL      bool
	monitScrapeURI string
	monitUser      string
	monitPassword  string
	logLevel       string
)

// RootCmd is the base command for this application.
var RootCmd = &cobra.Command{
	Use:   "monit_exporter",
	Short: "Monit Exporter for Prometheus",
	Long:  "Prometheus Exporter that collects Monit status information and exposes metrics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.Debug("RootCmd is running without subcommand, displaying help message")
		return cmd.Help()
	},
}

// Execute runs the root command of the application.
func Execute() {
	logrus.Debug("Execute function called: attempting to run RootCmd.Execute()")
	if err := RootCmd.Execute(); err != nil {
		logrus.Errorf("Error occurred while executing RootCmd: %v", err)
		fmt.Println(err)
		os.Exit(1)
	}
	logrus.Debug("Execute function finished: RootCmd.Execute() completed successfully")
}

func init() {
	RootCmd.PersistentFlags().StringVar(
		&listenAddress,
		"listen-address",
		"localhost:9388",
		"The address on which the exporter will listen (e.g., '0.0.0.0:9388').",
	)
	RootCmd.PersistentFlags().StringVar(
		&metricsPath,
		"metrics-path",
		"/metrics",
		"The HTTP path at which metrics are served (e.g., '/metrics').",
	)
	RootCmd.PersistentFlags().BoolVar(
		&ignoreSSL,
		"ignore-ssl",
		false,
		"Whether to skip SSL certificate verification for Monit endpoints.",
	)
	RootCmd.PersistentFlags().StringVar(
		&monitScrapeURI,
		"monit-scrape-uri",
		"http://localhost:2812/_status?format=xml&level=full",
		"The Monit status URL to scrape (XML format).",
	)
	RootCmd.PersistentFlags().StringVar(
		&monitUser,
		"monit-user",
		"",
		"Basic auth username for accessing Monit.",
	)
	RootCmd.PersistentFlags().StringVar(
		&monitPassword,
		"monit-password",
		"",
		"Basic auth password for accessing Monit.",
	)
	RootCmd.PersistentFlags().StringVar(
		&logLevel,
		"log-level",
		"info",
		"Log level for the application (debug, info, warn, error, fatal, panic).",
	)
}
