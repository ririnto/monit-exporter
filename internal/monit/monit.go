package monit

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ririnto/monit_exporter/internal/config"
	"golang.org/x/net/html/charset"
)

// XML represents the top-level structure of the Monit status XML.
// It includes <server>, <platform>, and multiple <service> elements.
type XML struct {
	XMLName  xml.Name  `xml:"monit"`
	Server   Server    `xml:"server"`
	Platform Platform  `xml:"platform"`
	Services []Service `xml:"service"`
}

// Server contains details from the <server> element in the Monit XML.
type Server struct {
	ID          string `xml:"id"`
	Incarnation int    `xml:"incarnation"`
	Version     string `xml:"version"`
	Uptime      int    `xml:"uptime"`
	Poll        int    `xml:"poll"`
	StartDelay  int    `xml:"startdelay"`
	LocalHost   string `xml:"localhostname"`
	ControlFile string `xml:"controlfile"`
	HTTPD       HTTPD  `xml:"httpd"`
}

// HTTPD represents the <httpd> element inside the <server> element.
type HTTPD struct {
	Address string `xml:"address"`
	Port    int    `xml:"port"`
	SSL     int    `xml:"ssl"`
}

// Platform represents the <platform> element in the Monit XML.
type Platform struct {
	Name    string `xml:"name"`
	Release string `xml:"release"`
	Version string `xml:"version"`
	Machine string `xml:"machine"`
	CPU     int    `xml:"cpu"`
	Memory  int    `xml:"memory"`
	Swap    int    `xml:"swap"`
}

// Service represents information about a single Monit service.
// It is mapped from the <service> element (including the "type" attribute).
type Service struct {
	Type      int    `xml:"type,attr"`
	Name      string `xml:"name"`
	Status    int    `xml:"status"`
	Monitored string `xml:"monitor"`
}

// FetchMonitStatus sends an HTTP GET request to the Monit endpoint and
// returns the response body. It applies a 5-second timeout context to
// prevent indefinite waiting.
func FetchMonitStatus(cfg *config.Config) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", cfg.MonitScrapeURI, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	req.SetBasicAuth(cfg.MonitUser, cfg.MonitPassword)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.IgnoreSSL},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch Monit status: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("monit returned non-2xx status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read Monit status: %w", err)
	}

	return data, nil
}

// ParseMonitStatus parses the Monit XML data into an XML struct, including
// fields for <server>, <platform>, and <service> elements.
func ParseMonitStatus(data []byte) (XML, error) {
	var statusChunk XML
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(&statusChunk); err != nil {
		return XML{}, fmt.Errorf("failed to parse Monit XML: %w", err)
	}
	return statusChunk, nil
}
