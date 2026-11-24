package cmd

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "monit-exporter.exe")
	if output, err := exec.Command("go", "build", "-o", binary, "..").CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	const password = "monit-debug-test-password-7c43b1"
	for _, test := range []struct {
		name     string
		args     []string
		exitCode int
		want     []string
	}{
		{
			name: "root help",
			want: []string{"monit-exporter [command]", "Run the Monit Exporter server"},
		},
		{
			name: "serve help",
			args: []string{"serve", "--help"},
			want: []string{
				"monit-exporter serve [flags]",
				"--ignore-ssl",
				"--listen-address string", `(default "localhost:9388")`,
				"--log-level string", `(default "info")`,
				"--metrics-path string", `(default "/metrics")`,
				"--monit-password string",
				"--monit-scrape-uri string", `(default "http://localhost:2812/_status?format=xml&level=full")`,
				"--monit-user string",
			},
		},
		{
			name:     "unknown flag",
			args:     []string{"serve", "--unknown-flag"},
			exitCode: 1,
			want:     []string{"unknown flag: --unknown-flag"},
		},
		{
			name: "debug password",
			args: []string{
				"serve", "--log-level", "debug",
				"--listen-address", "127.0.0.1:65536",
				"--metrics-path", "/compat-metrics",
				"--monit-user", "monit-test-user",
				"--monit-password", password,
			},
			exitCode: 1,
			want:     []string{"level=debug", "/compat-metrics", "failed to start server"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			output, err := exec.CommandContext(ctx, binary, test.args...).CombinedOutput()
			var exitErr *exec.ExitError
			exitCode := 0
			if errors.As(err, &exitErr) {
				exitCode = exitErr.ExitCode()
			} else if err != nil {
				t.Fatalf("run CLI: %v", err)
			}
			if strings.Contains(string(output), password) {
				t.Error("CLI output contains the configured Monit password")
			}
			if exitCode != test.exitCode {
				t.Errorf("CLI exit code = %d, want %d", exitCode, test.exitCode)
			}
			for _, want := range test.want {
				if !strings.Contains(string(output), want) {
					t.Errorf("CLI output does not contain %q", want)
				}
			}
		})
	}
}
