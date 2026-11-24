package cmd

import (
	"embed"
	"fmt"
	"strings"

	"github.com/alecthomas/kong"
)

//go:embed completion/*
var completionScripts embed.FS

type completionCommand struct {
	Bash       completionScript `cmd:"" help:"Generate the autocompletion script for bash"`
	Fish       completionScript `cmd:"" help:"Generate the autocompletion script for fish"`
	PowerShell completionScript `cmd:"" name:"powershell" help:"Generate the autocompletion script for powershell"`
	Zsh        completionScript `cmd:"" help:"Generate the autocompletion script for zsh"`
}

func (*completionCommand) Help() string {
	return "Generate the autocompletion script for monit-exporter for the specified shell.\n" +
		"See each sub-command's help for details on how to use the generated script."
}

func (*completionCommand) Run(ctx *kong.Context) error {
	if ctx.Selected().Parent == ctx.Model.Node {
		return printHelp(ctx.Selected(), ctx.Stdout)
	}
	return nil
}

var completionSetup = map[string]string{
	"bash":       "source <(monit-exporter completion bash)",
	"fish":       "monit-exporter completion fish | source",
	"powershell": "monit-exporter completion powershell | Out-String | Invoke-Expression",
	"zsh":        "autoload -Uz compinit && compinit; source <(monit-exporter completion zsh)",
}

type completionScript struct {
	NoDescriptions boolFlag `help:"Disable completion descriptions."`
}

func (c *completionScript) Run(ctx *kong.Context) error {
	script, err := completionScripts.ReadFile("completion/" + ctx.Selected().Name)
	if err != nil {
		return err
	}
	callback := "__complete"
	if c.NoDescriptions {
		callback = "__completeNoDesc"
	}
	_, err = fmt.Fprint(ctx.Stdout, strings.ReplaceAll(string(script), "__COMPLETE_COMMAND__", callback))
	return err
}

type completeRequest struct {
	Args []string `arg:"" name:"command-line" sep:"none"`
}

func (c *completeRequest) Run(ctx *kong.Context) error {
	prefix := c.Args[len(c.Args)-1]
	trace, err := kong.Trace(ctx.Kong, c.Args[:len(c.Args)-1])
	if err != nil {
		return err
	}

	var rows strings.Builder
	writeCandidate := func(name, description string) {
		if !strings.HasPrefix(name, prefix) {
			return
		}
		rows.WriteString(name)
		if ctx.Selected().Name == "__complete" {
			fmt.Fprintf(&rows, "\t%s", description)
		}
		rows.WriteByte('\n')
	}

	const (
		allowFileCompletion = 0
		noFileCompletion    = 4
	)
	directive := allowFileCompletion
	if trace.Error == nil {
		node := selectedCommand(trace)
		if node.Name == "help" {
			node = ctx.Model.Node
		}
		if strings.HasPrefix(prefix, "-") {
			directive = noFileCompletion
			used := make(map[*kong.Flag]bool)
			for _, path := range trace.Path {
				used[path.Flag] = true
			}
			for _, group := range node.AllFlags(true) {
				for _, flag := range group {
					if !used[flag] {
						help := flag.Help
						if flag.Name == "help" {
							help = "help for " + node.Name
						}
						writeCandidate("--"+flag.Name, help)
						if flag.Short != 0 {
							writeCandidate("-"+string(flag.Short), help)
						}
					}
				}
			}
		} else if len(node.Children) > 0 {
			directive = noFileCompletion
			for _, command := range node.Children {
				if !command.Hidden {
					writeCandidate(command.Name, command.Help)
				}
			}
		}
	}
	fmt.Fprintf(&rows, ":%d\n", directive)
	_, err = fmt.Fprint(ctx.Stdout, rows.String())
	return err
}
