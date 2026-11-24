package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/alecthomas/kong"
)

type helpCommand struct {
	Topic []string `arg:"" optional:"" sep:"none" help:"Command to show help for."`
}

func (*helpCommand) Help() string {
	return "Help provides help for any command in the application.\nUse monit-exporter help [command] for more information."
}

func (h *helpCommand) Run(ctx *kong.Context) error {
	topic, err := kong.Trace(ctx.Kong, h.Topic)
	if err != nil {
		return err
	}
	if topic.Error != nil {
		if _, err := fmt.Fprintf(ctx.Stderr, "Unknown help topic %q\n", strings.Join(h.Topic, " ")); err != nil {
			return err
		}
		return printHelp(ctx.Model.Node, ctx.Stderr)
	}
	return printHelp(selectedCommand(topic), ctx.Stdout)
}

func selectedCommand(ctx *kong.Context) *kong.Node {
	if node := ctx.Selected(); node != nil {
		return node
	}
	return ctx.Model.Node
}

func printHelp(node *kong.Node, writer io.Writer) error {
	var output bytes.Buffer
	desc := node.Help
	if node.Detail != "" {
		desc = node.Detail
	}
	fmt.Fprintf(&output, "%s\n", desc)
	if node.Parent != nil && node.Parent.Name == "completion" {
		fmt.Fprintf(&output, "\nTo load completions in your current shell session:\n\n\t%s\n", completionSetup[node.Name])
	}
	fmt.Fprintln(&output, "\nUsage:")
	path := node.FullPath()
	if node.Parent == nil {
		fmt.Fprintf(&output, "  %s [flags]\n", path)
	}
	switch {
	case len(node.Children) > 0:
		fmt.Fprintf(&output, "  %s [command]\n\nAvailable Commands:\n", path)
		columns := tabwriter.NewWriter(&output, 0, 0, 2, ' ', 0)
		for _, command := range node.Children {
			if !command.Hidden {
				_, _ = fmt.Fprintf(columns, "  %s\t%s\n", command.Name, command.Help)
			}
		}
		_ = columns.Flush()
	case node.Name == "help":
		fmt.Fprintf(&output, "  %s [command] [flags]\n", path)
	default:
		fmt.Fprintf(&output, "  %s [flags]\n", path)
	}

	fmt.Fprintln(&output, "\nFlags:")
	columns := tabwriter.NewWriter(&output, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintf(columns, "  -h, --help[=true|false]\thelp for %s\n", node.Name)
	for _, flag := range node.Flags {
		if flag.Name != "help" {
			writeFlagHelp(columns, flag)
		}
	}
	_ = columns.Flush()
	if node.Parent != nil {
		fmt.Fprintln(&output, "\nGlobal Flags:")
		for _, group := range node.Parent.AllFlags(true) {
			for _, flag := range group {
				if flag.Name != "help" {
					writeFlagHelp(columns, flag)
				}
			}
		}
		_ = columns.Flush()
	}

	if len(node.Children) > 0 {
		fmt.Fprintf(&output, "\nUse %q for more information about a command.\n", path+" [command] --help")
	}
	_, err := writer.Write(output.Bytes())
	return err
}

func writeFlagHelp(writer io.Writer, flag *kong.Flag) {
	suffix := " string"
	if flag.IsBool() {
		suffix = "[=true|false]"
	}
	_, _ = fmt.Fprintf(writer, "      --%s%s\t%s", flag.Name, suffix, flag.Help)
	if flag.HasDefault {
		_, _ = fmt.Fprintf(writer, " (default %q)", flag.Default)
	}
	_, _ = fmt.Fprintln(writer)
}
