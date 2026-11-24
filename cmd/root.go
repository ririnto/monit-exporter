// Package cmd defines the Monit Exporter command-line interface.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

type commandLine struct {
	Help           boolFlag `short:"h" help:"Show help for the current command."`
	IgnoreSSL      boolFlag `name:"ignore-ssl" help:"Whether to skip SSL certificate verification for Monit endpoints."`
	ListenAddress  string   `default:"localhost:9388" help:"The address on which the exporter will listen (e.g., '0.0.0.0:9388')."`
	LogLevel       string   `default:"info" help:"Log level for the application (debug, info, warn, error, fatal, panic)."`
	MetricsPath    string   `default:"/metrics" help:"The HTTP path at which metrics are served (e.g., '/metrics')."`
	MonitPassword  string   `help:"Basic auth password for accessing Monit."`
	MonitScrapeURI string   `name:"monit-scrape-uri" default:"http://localhost:2812/_status?format=xml&level=full" help:"The Monit status URL to scrape (XML format)."`
	MonitUser      string   `help:"Basic auth username for accessing Monit."`

	Completion     completionCommand `cmd:"" help:"Generate the autocompletion script for the specified shell"`
	HelpTopic      helpCommand       `cmd:"" name:"help" help:"Help about any command"`
	Serve          serveCommand      `cmd:"" help:"Run the Monit Exporter server"`
	Complete       completeRequest   `cmd:"" name:"__complete" hidden:"" passthrough:""`
	CompleteNoDesc completeRequest   `cmd:"" name:"__completeNoDesc" hidden:"" passthrough:""`
}

func (*commandLine) Run(ctx *kong.Context) error {
	if ctx.Selected() == nil {
		return printHelp(ctx.Model.Node, ctx.Stdout)
	}
	return nil
}

type boolFlag bool

func (b *boolFlag) Decode(ctx *kong.DecodeContext) error {
	token := ctx.Scan.Peek()
	switch {
	case token.Type == kong.FlagValueToken:
		ctx.Scan.Pop()
	case token.Type == kong.ShortFlagTailToken && strings.HasPrefix(token.String(), "="):
		ctx.Scan.Pop()
		token.Value = strings.TrimPrefix(token.String(), "=")
	default:
		*b = true
		return nil
	}

	value, err := strconv.ParseBool(token.String())
	if err != nil {
		return fmt.Errorf("invalid boolean %q (must be true or false)", token.String())
	}
	*b = boolFlag(value)
	return nil
}

// Execute parses the command line and exits with status 1 on failure.
func Execute() {
	var cli commandLine
	parser, err := kong.New(&cli,
		kong.Name("monit-exporter"),
		kong.Description("Prometheus Exporter that collects Monit status information and exposes metrics."),
		kong.NoDefaultHelp(),
		kong.KindMapper(reflect.String, kong.MapperFunc(func(ctx *kong.DecodeContext, target reflect.Value) error {
			token := ctx.Scan.Pop()
			if token.IsEOL() {
				return errors.New("flag needs an argument")
			}
			target.SetString(token.String())
			return nil
		})),
	)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	ctx, err := parser.Parse(os.Args[1:])
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		ctx = parseErr.Context
	}
	unexpectedArg := err != nil && strings.HasPrefix(err.Error(), "unexpected argument ")
	if ctx != nil && (err == nil || unexpectedArg) {
		showHelp := unexpectedArg && selectedCommand(ctx).Name == "completion"
		for i, path := range ctx.Path {
			if path.Flag != nil && path.Flag.Name == "help" && ctx.FlagValue(path.Flag) == boolFlag(true) {
				ctx.Path = ctx.Path[:i]
				showHelp = true
				break
			}
		}
		if showHelp {
			err = printHelp(selectedCommand(ctx), ctx.Stdout)
		} else if err == nil {
			err = ctx.Run()
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if ctx != nil {
			_ = printHelp(selectedCommand(ctx), os.Stderr)
		}
		logrus.Error(err)
		fmt.Println(err)
		os.Exit(1)
	}
}
