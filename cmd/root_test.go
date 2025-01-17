package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

// TestRootCmd_Help verifies that running RootCmd without subcommands prints help.
func TestRootCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{})

	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("RootCmd execution failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("Expected help output, got empty string")
	}
}

// TestRootCmd_Execute checks if Execute() runs RootCmd properly.
func TestRootCmd_Execute(t *testing.T) {
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test subcommand",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	RootCmd.AddCommand(testCmd)
	defer RootCmd.RemoveCommand(testCmd)

	RootCmd.SetArgs([]string{"test"})
	Execute()
}
