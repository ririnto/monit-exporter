package monit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ririnto/monit-exporter/internal/config"
)

func TestRedactRequestURL_MalformedRedirect(t *testing.T) {
	const location = "https://redirect.example.invalid:location-marker"
	requestURL := url.URL{Scheme: "https", Host: "monit.example.invalid", Path: "/request-marker"}
	_, parseErr := requestURL.Parse(location)
	if parseErr == nil {
		t.Fatal("Expected the synthetic Location to be malformed")
	}
	err := redactRequestURL(&url.Error{
		Op:  "Get",
		URL: requestURL.String(),
		Err: fmt.Errorf("failed to parse Location header %q: %v", location, parseErr),
	})
	message := err.Error()
	if strings.Contains(message, "request-marker") {
		t.Error("The request URL still appears in the error")
	}
	if strings.Contains(message, "location-marker") {
		t.Error("The malformed Location still appears in the error")
	}
	if !strings.Contains(message, "Get") || !strings.Contains(message, "failed to parse Location header") {
		t.Errorf("Expected the HTTP operation and redirect parse diagnosis, got %q", message)
	}
}

func TestRedactRequestURL_PreservesOtherCause(t *testing.T) {
	err := redactRequestURL(&url.Error{
		Op:  "Get",
		URL: "https://monit.example.invalid/request-marker",
		Err: context.DeadlineExceeded,
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected the timeout cause to remain available, got %v", err)
	}
}

// TestFetchMonitStatus_Success checks if FetchMonitStatus can retrieve mock XML successfully.
func TestFetchMonitStatus_Success(t *testing.T) {
	t.Log("Testing FetchMonitStatus with a mock server providing valid XML")

	mockXML := `<?xml version="1.0"?><monit><server><version>5.26.0</version></server></monit>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, mockXML)
	}))
	defer server.Close()

	cfg := &config.Config{MonitScrapeURI: server.URL, IgnoreSSL: false}
	data, err := FetchMonitStatus(cfg)
	if err != nil {
		t.Fatalf("FetchMonitStatus returned error: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("Expected non-empty data, got empty")
	}
}

// TestFetchMonitStatus_Non2xx checks if FetchMonitStatus returns an error on 400 status code.
func TestFetchMonitStatus_Non2xx(t *testing.T) {
	t.Log("Testing FetchMonitStatus with a mock server returning 400 status")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "some error", http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := &config.Config{MonitScrapeURI: server.URL}
	_, err := FetchMonitStatus(cfg)
	if err == nil {
		t.Fatal("Expected an error for HTTP 400 status, got nil")
	}
}

// TestParseMonitStatus_Success verifies parsing a valid Monit XML.
func TestParseMonitStatus_Success(t *testing.T) {
	t.Log("Testing ParseMonitStatus with a valid XML string")

	mockXML := `<?xml version="1.0"?><monit><server><version>5.26.0</version></server><platform/><service type="0"><name>rootfs</name></service></monit>`
	monitData, err := ParseMonitStatus([]byte(mockXML))
	if err != nil {
		t.Fatalf("ParseMonitStatus failed: %v", err)
	}

	if len(monitData.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(monitData.Services))
	}
	if monitData.Services[0].Name != "rootfs" {
		t.Errorf("Expected service name 'rootfs', got '%s'", monitData.Services[0].Name)
	}
}

// TestParseMonitStatus_Error verifies error handling for malformed XML.
func TestParseMonitStatus_Error(t *testing.T) {
	t.Log("Testing ParseMonitStatus with malformed XML string")

	invalidXML := `<monit><server>missing closing tags`
	_, err := ParseMonitStatus([]byte(invalidXML))
	if err == nil {
		t.Fatal("Expected an XML parse error, got nil")
	}
}
