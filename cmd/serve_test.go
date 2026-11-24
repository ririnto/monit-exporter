package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCommonLogHandler(t *testing.T) {
	var logs bytes.Buffer
	defer logrus.SetOutput(logrus.StandardLogger().Out)
	logrus.SetOutput(&logs)

	const body = "OK"
	handler := commonLogHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		if n, err := w.Write([]byte(body)); n != len(body) || err != nil {
			t.Errorf("Write returned (%d, %v), want (%d, nil)", n, err, len(body))
		}
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/dummy", nil))
	if got := response.Result().StatusCode; got != http.StatusAccepted {
		t.Errorf("HTTP status = %d, want %d", got, http.StatusAccepted)
	}
	if got := response.Body.String(); got != body {
		t.Errorf("HTTP body = %q, want %q", got, body)
	}
	if got := logs.String(); !strings.Contains(got, " 202 2 ") {
		t.Errorf("access log = %q, want status 202 and size 2", got)
	}
}
