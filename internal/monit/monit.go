package monit

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
	"time"

	"github.com/commercetools/monit-exporter/internal/config"
	"github.com/sirupsen/logrus"
)

// Monit represents the top-level XML element <monit>.
type Monit struct {
	XMLName  xml.Name  `xml:"monit"`
	Server   Server    `xml:"server"`
	Platform Platform  `xml:"platform"`
	Services []Service `xml:"service"`
}

// Server represents the <server> element in the Monit XML.
type Server struct {
	ID            string `xml:"id"`
	Incarnation   int64  `xml:"incarnation"`
	Version       string `xml:"version"`
	Uptime        int64  `xml:"uptime"`
	Poll          int    `xml:"poll"`
	StartDelay    int    `xml:"startdelay"`
	Localhostname string `xml:"localhostname"`
	Controlfile   string `xml:"controlfile"`
	HTTPD         HTTPD  `xml:"httpd"`
}

// HTTPD represents the <httpd> element in the Monit XML.
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
	Memory  int64  `xml:"memory"`
	Swap    int64  `xml:"swap"`
}

// Service represents the <service> element in the Monit XML.
type Service struct {
	Type          int     `xml:"type,attr"`
	Name          string  `xml:"name"`
	CollectedSec  int64   `xml:"collected_sec"`
	CollectedUsec int64   `xml:"collected_usec"`
	Status        int     `xml:"status"`
	StatusHint    int     `xml:"status_hint"`
	Monitor       int     `xml:"monitor"`
	MonitorMode   int     `xml:"monitormode"`
	OnReboot      int     `xml:"onreboot"`
	PendingAction int     `xml:"pendingaction"`
	Fstype        string  `xml:"fstype,omitempty"`
	Fsflags       string  `xml:"fsflags,omitempty"`
	Mode          string  `xml:"mode,omitempty"`
	UID           int     `xml:"uid,omitempty"`
	GID           int     `xml:"gid,omitempty"`
	Block         *Block  `xml:"block,omitempty"`
	Inode         *Inode  `xml:"inode,omitempty"`
	Read          string  `xml:"read,omitempty"`
	Write         string  `xml:"write,omitempty"`
	Port          *Port   `xml:"port,omitempty"`
	System        *System `xml:"system,omitempty"`
	Link          *Link   `xml:"link,omitempty"`
}

// Block represents the <block> element under a filesystem service.
type Block struct {
	Percent float64 `xml:"percent"`
	Usage   float64 `xml:"usage"`
	Total   float64 `xml:"total"`
}

// Inode represents the <inode> element under a filesystem service.
type Inode struct {
	Percent float64 `xml:"percent"`
	Usage   int     `xml:"usage"`
	Total   int     `xml:"total"`
}

// Port represents the <port> element, typically for remote host checks.
type Port struct {
	Hostname     string      `xml:"hostname"`
	Portnumber   int         `xml:"portnumber"`
	Request      string      `xml:"request"`
	Protocol     string      `xml:"protocol"`
	Type         string      `xml:"type"`
	Responsetime float64     `xml:"responsetime"`
	Certificate  Certificate `xml:"certificate"`
}

// Certificate represents the <certificate> element under <port>.
type Certificate struct {
	Valid int `xml:"valid"`
}

// System represents the <system> element, usually present in type="5" (System) services.
type System struct {
	Load   Load   `xml:"load"`
	CPU    CPU    `xml:"cpu"`
	Memory Memory `xml:"memory"`
	Swap   Swap   `xml:"swap"`
}

// Load represents the <load> element under <system>.
type Load struct {
	Avg01 float64 `xml:"avg01"`
	Avg05 float64 `xml:"avg05"`
	Avg15 float64 `xml:"avg15"`
}

// CPU represents the <cpu> element under <system>.
type CPU struct {
	User   float64 `xml:"user"`
	System float64 `xml:"system"`
	Wait   float64 `xml:"wait"`
}

// Memory represents the <memory> element under <system>.
type Memory struct {
	Percent  float64 `xml:"percent"`
	Kilobyte int     `xml:"kilobyte"`
}

// Swap represents the <swap> element under <system>.
type Swap struct {
	Percent  float64 `xml:"percent"`
	Kilobyte int     `xml:"kilobyte"`
}

// Link represents the <link> element under a network service.
type Link struct {
	State    int      `xml:"state"`
	Speed    int64    `xml:"speed"`
	Duplex   int      `xml:"duplex"`
	Download Download `xml:"download"`
	Upload   Upload   `xml:"upload"`
}

// Download represents the <download> element under <link>.
type Download struct {
	Packets Packets `xml:"packets"`
	Bytes   Bytes   `xml:"bytes"`
	Errors  Errors  `xml:"errors"`
}

// Upload represents the <upload> element under <link>.
type Upload struct {
	Packets Packets `xml:"packets"`
	Bytes   Bytes   `xml:"bytes"`
	Errors  Errors  `xml:"errors"`
}

// Packets represents the <packets> element under <download> or <upload>.
type Packets struct {
	Now   int `xml:"now"`
	Total int `xml:"total"`
}

// Bytes represents the <bytes> element under <download> or <upload>.
type Bytes struct {
	Now   int `xml:"now"`
	Total int `xml:"total"`
}

// Errors represents the <errors> element under <download> or <upload>.
type Errors struct {
	Now   int `xml:"now"`
	Total int `xml:"total"`
}

// FetchMonitStatus sends an HTTP GET request to the Monit endpoint and returns the response body.
func FetchMonitStatus(cfg *config.Config) ([]byte, error) {
	logrus.Debugf("FetchMonitStatus: MonitScrapeURI=%s, IgnoreSSL=%t", cfg.MonitScrapeURI, cfg.IgnoreSSL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", cfg.MonitScrapeURI, nil)
	if err != nil {
		logrus.Errorf("FetchMonitStatus: failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("unable to create request: %w", err)
	}
	req.SetBasicAuth(cfg.MonitUser, cfg.MonitPassword)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.IgnoreSSL},
	}
	client := &http.Client{Transport: tr}

	logrus.Debug("FetchMonitStatus: sending request to Monit")
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("FetchMonitStatus: HTTP request failed: %v", err)
		return nil, fmt.Errorf("unable to fetch Monit status: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logrus.Warnf("FetchMonitStatus: failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logrus.Errorf("FetchMonitStatus: non-2xx status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("monit returned non-2xx status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("FetchMonitStatus: failed to read response body: %v", err)
		return nil, fmt.Errorf("unable to read Monit status: %w", err)
	}
	logrus.Debugf("FetchMonitStatus: successfully received response (%d bytes)", len(data))
	return data, nil
}

// ParseMonitStatus parses the XML data and returns a Monit struct.
func ParseMonitStatus(data []byte) (Monit, error) {
	logrus.Debug("ParseMonitStatus: starting XML parsing")
	var statusChunk Monit
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(&statusChunk); err != nil {
		logrus.Errorf("ParseMonitStatus: XML parsing failed: %v", err)
		return Monit{}, fmt.Errorf("failed to parse Monit XML: %w", err)
	}
	logrus.Debugf("ParseMonitStatus: successfully parsed. Services count=%d", len(statusChunk.Services))
	return statusChunk, nil
}
