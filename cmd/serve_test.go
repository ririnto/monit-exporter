package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"serve", "--help"})

	err := serveCmd.Execute()
	if err != nil {
		t.Fatalf("ServeCmd execution failed with --help: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Errorf("Expected help output for serve command, got empty string")
	}
}

func TestCommonLogHandler(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	handler := commonLogHandler(dummyHandler)

	req := httptest.NewRequest("GET", "/dummy", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}
}

func TestNewLoggingResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	lrw := NewLoggingResponseWriter(rec)
	lrw.WriteHeader(http.StatusAccepted)

	if lrw.statusCode != http.StatusAccepted {
		t.Errorf("Expected status code 202, got %d", lrw.statusCode)
	}

	testData := []byte("Hello, World!")
	n, err := lrw.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected %d bytes, wrote %d", len(testData), n)
	}
	if lrw.size != len(testData) {
		t.Errorf("Expected lrw.size=%d, got %d", len(testData), lrw.size)
	}
}
