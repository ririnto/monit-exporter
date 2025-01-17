package monit

import (
	"testing"
)

// TestParseMonitStatusMinimal checks parsing with a minimal Monit XML snippet.
func TestParseMonitStatusMinimal(t *testing.T) {
	xmlData := []byte(`
<monit>
  <service type="5">
    <name>test_service</name>
    <status>0</status>
    <monitor>1</monitor>
  </service>
</monit>
`)
	parsed, err := ParseMonitStatus(xmlData)
	if err != nil {
		t.Fatalf("ParseMonitStatus failed: %v", err)
	}
	if len(parsed.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(parsed.Services))
	}

	svc := parsed.Services[0]
	if svc.Type != 5 {
		t.Errorf("expected type=5, got %d", svc.Type)
	}
	if svc.Name != "test_service" {
		t.Errorf("expected name=test_service, got %s", svc.Name)
	}
	if svc.Status != 0 {
		t.Errorf("expected status=0, got %d", svc.Status)
	}
	if svc.Monitored != "1" {
		t.Errorf("expected monitor=1, got %s", svc.Monitored)
	}
}

// TestParseMonitStatusFullXML checks parsing with the full Monit XML snippet
// that includes <server>, <platform>, and <service>.
func TestParseMonitStatusFullXML(t *testing.T) {
	xmlData := []byte(`
<?xml version="1.0" encoding="ISO-8859-1"?>
<monit>
  <server>
    <id>acfbb9e9118e68d3754761a79d3aae16</id>
    <incarnation>1504605214</incarnation>
    <version>5.23.0</version>
    <uptime>136736</uptime>
    <poll>60</poll>
    <startdelay>0</startdelay>
    <localhostname>fc566edc8b68</localhostname>
    <controlfile>/opt/monit/etc/monitrc</controlfile>
    <httpd>
      <address>172.17.0.2</address>
      <port>2812</port>
      <ssl>0</ssl>
    </httpd>
  </server>
  <platform>
    <name>Linux</name>
    <release>4.9.27-moby</release>
    <version>#1 SMP Thu May 11 04:01:18 UTC 2017</version>
    <machine>x86_64</machine>
    <cpu>4</cpu>
    <memory>2046768</memory>
    <swap>1048572</swap>
  </platform>
  <service type="5">
    <name>fc566edc8b68</name>
    <status>0</status>
    <monitor>1</monitor>
  </service>
</monit>
`)

	parsed, err := ParseMonitStatus(xmlData)
	if err != nil {
		t.Fatalf("ParseMonitStatus failed: %v", err)
	}

	// Check <server>
	if parsed.Server.ID != "acfbb9e9118e68d3754761a79d3aae16" {
		t.Errorf("expected Server.ID=acfbb9e9118e68d3754761a79d3aae16, got %s", parsed.Server.ID)
	}
	if parsed.Server.Uptime != 136736 {
		t.Errorf("expected Server.Uptime=136736, got %d", parsed.Server.Uptime)
	}
	if parsed.Server.HTTPD.Port != 2812 {
		t.Errorf("expected Server.HTTPD.Port=2812, got %d", parsed.Server.HTTPD.Port)
	}

	// Check <platform>
	if parsed.Platform.Name != "Linux" {
		t.Errorf("expected platform name=Linux, got %s", parsed.Platform.Name)
	}
	if parsed.Platform.Release != "4.9.27-moby" {
		t.Errorf("expected platform release=4.9.27-moby, got %s", parsed.Platform.Release)
	}
	if parsed.Platform.CPU != 4 {
		t.Errorf("expected CPU=4, got %d", parsed.Platform.CPU)
	}

	// Check <service>
	if len(parsed.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(parsed.Services))
		return
	}
	svc := parsed.Services[0]
	if svc.Type != 5 {
		t.Errorf("expected service type=5, got %d", svc.Type)
	}
	if svc.Name != "fc566edc8b68" {
		t.Errorf("expected service name=fc566edc8b68, got %s", svc.Name)
	}
	if svc.Status != 0 {
		t.Errorf("expected status=0, got %d", svc.Status)
	}
	if svc.Monitored != "1" {
		t.Errorf("expected monitored=1, got %s", svc.Monitored)
	}
}

// TestParseMonitStatusInvalidXML checks behavior with invalid XML.
func TestParseMonitStatusInvalidXML(t *testing.T) {
	xmlData := []byte(`<<<invalid xml>>>`)
	_, err := ParseMonitStatus(xmlData)
	if err == nil {
		t.Errorf("expected parsing error but got nil")
	}
}
