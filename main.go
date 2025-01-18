package main

import (
	"github.com/ririnto/monit-exporter/cmd"
	"github.com/sirupsen/logrus"
)

// main is the entry point of the Monit Exporter application.
func main() {
	logrus.Debug("main() function invoked: calling cmd.Execute()")
	cmd.Execute()
	logrus.Debug("main() function completed: cmd.Execute() returned without error")
}
