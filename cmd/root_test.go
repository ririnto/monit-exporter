package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCmd checks if RootCmd is a valid cobra.Command.
func TestRootCmd(t *testing.T) {
	var _ *cobra.Command = RootCmd
}
