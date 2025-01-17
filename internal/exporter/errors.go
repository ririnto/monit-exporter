package exporter

import "errors"

// ErrNilConfig is returned when a nil config is provided to NewExporter.
var ErrNilConfig = errors.New("config is nil")
