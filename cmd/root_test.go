package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
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
				"--ignore-ssl[=true|false]",
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

	t.Run("debug scrape URI credentials", func(t *testing.T) {
		const (
			uriUser     = "monit-uri-test-user-1b92d4"
			uriPassword = "monit-uri-test-password-5e78c2"
			uriToken    = "monit-uri-test-token-9ac41f"
			query       = "format=xml&token=" + uriToken
		)
		monit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawQuery != query {
				http.Error(w, "unexpected Monit request", http.StatusBadRequest)
				return
			}
			switch r.URL.Path {
			case "/_status":
				_, _ = io.WriteString(w, `<?xml version="1.0"?><monit><server><version>5.26.0</version></server></monit>`)
			case "/abort":
				panic(http.ErrAbortHandler)
			default:
				http.Error(w, "unexpected Monit path", http.StatusBadRequest)
			}
		}))
		defer monit.Close()
		scrapeURI := url.URL{
			Scheme: "http",
			Host:   monit.Listener.Addr().String(),
			User:   url.UserPassword(uriUser, uriPassword),
		}

		for _, test := range []struct {
			name      string
			path      string
			wantUp    string
			wantCause string
		}{
			{name: "success", path: "/_status", wantUp: "1"},
			{name: "transport failure", path: "/abort", wantUp: "0", wantCause: "EOF"},
			{name: "invalid URI", path: "/_status%zz", wantUp: "0", wantCause: "invalid URL escape"},
		} {
			t.Run(test.name, func(t *testing.T) {
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatalf("allocate exporter address: %v", err)
				}
				address := listener.Addr().String()
				if err := listener.Close(); err != nil {
					t.Fatalf("release exporter address: %v", err)
				}

				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
				defer cancel()
				var output bytes.Buffer
				command := exec.CommandContext(ctx, binary, "serve", "--log-level", "debug",
					"--listen-address", address, "--monit-scrape-uri", scrapeURI.String()+test.path+"?"+query)
				command.Stdout, command.Stderr = &output, &output
				if err := command.Start(); err != nil {
					t.Fatalf("start CLI: %v", err)
				}
				defer func() {
					cancel()
					_ = command.Wait()
					if !strings.Contains(output.String(), "level=debug") {
						t.Error("CLI output does not contain debug logs")
					}
					if test.wantCause != "" && !strings.Contains(output.String(), test.wantCause) {
						t.Errorf("CLI output does not contain the scrape failure cause %q", test.wantCause)
					}
					for field, value := range map[string]string{
						"username":    uriUser,
						"password":    uriPassword,
						"query token": uriToken,
					} {
						if strings.Contains(output.String(), value) {
							t.Errorf("CLI output contains the Monit scrape URI %s", field)
						}
					}
				}()

				request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address+"/metrics", nil)
				if err != nil {
					t.Fatalf("create metrics request: %v", err)
				}
				request.Header.Set("Accept", "text/plain;version=1.0.0")
				request.Header.Set("Accept-Encoding", "zstd")
				for {
					response, err := http.DefaultClient.Do(request)
					if err == nil {
						defer func() {
							_ = response.Body.Close()
						}()
						if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain; version=0.0.4;") {
							t.Fatalf("metrics content type = %q, want Prometheus text 0.0.4", contentType)
						}
						if encoding := response.Header.Get("Content-Encoding"); encoding != "zstd" {
							t.Fatalf("metrics content encoding = %q, want zstd", encoding)
						}
						decoder, err := zstd.NewReader(response.Body)
						if err != nil {
							t.Fatalf("create metrics decoder: %v", err)
						}
						defer decoder.Close()
						metrics, err := io.ReadAll(decoder)
						if err != nil {
							t.Fatalf("read metrics: %v", err)
						}
						if response.StatusCode != http.StatusOK || !strings.Contains(string(metrics), "monit_exporter_up "+test.wantUp+"\n") {
							t.Fatalf("expected monit_exporter_up %s, metrics response status %d", test.wantUp, response.StatusCode)
						}
						if runtime.GOOS == "darwin" && strings.Contains(string(metrics), "\n# TYPE process_") {
							t.Error("metrics include process metrics on macOS")
						}
						break
					}
					if ctx.Err() != nil {
						t.Fatalf("request CLI metrics: %v", err)
					}
					time.Sleep(10 * time.Millisecond)
				}
			})
		}
	})
}
