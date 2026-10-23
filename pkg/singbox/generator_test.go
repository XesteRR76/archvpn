package singbox_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"archvpn/pkg/node"
	"archvpn/pkg/singbox"
)

func TestGenerateConfig(t *testing.T) {
	n := &node.Node{
		ID:          "test-node",
		Protocol:    "vless",
		UUID:        "00000000-0000-0000-0000-000000000001",
		Server:      "192.0.2.1",
		Port:        443,
		Flow:        "xtls-rprx-vision",
		Security:    "reality",
		SNI:         "example.com",
		Fingerprint: "chrome",
		PublicKey:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		ShortID:     "1a2b3c4d",
		NetworkType: "tcp",
		Remark:      "TestServer",
	}

	data, err := singbox.GenerateConfig(n, "tun0", []string{"1.1.1.1", "8.8.8.8"})
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid json generated: %v", err)
	}

	inbounds, ok := parsed["inbounds"].([]interface{})
	if !ok || len(inbounds) == 0 {
		t.Fatalf("missing inbounds")
	}

	tunInbound := inbounds[0].(map[string]interface{})
	if tunInbound["type"] != "tun" {
		t.Errorf("expected inbound type tun, got %v", tunInbound["type"])
	}
	if tunInbound["interface_name"] != "tun0" {
		t.Errorf("expected interface tun0, got %v", tunInbound["interface_name"])
	}

	outbounds, ok := parsed["outbounds"].([]interface{})
	if !ok || len(outbounds) == 0 {
		t.Fatalf("missing outbounds")
	}

	proxyOutbound := outbounds[0].(map[string]interface{})
	if proxyOutbound["type"] != "vless" {
		t.Errorf("expected proxy outbound type vless, got %v", proxyOutbound["type"])
	}
	if proxyOutbound["server"] != "192.0.2.1" {
		t.Errorf("expected server 192.0.2.1, got %v", proxyOutbound["server"])
	}
}

func TestGenerateConfig_SingboxCheck(t *testing.T) {
	n := &node.Node{
		ID:          "test-node-2",
		Protocol:    "vless",
		UUID:        "00000000-0000-0000-0000-000000000002",
		Server:      "192.0.2.2",
		Port:        443,
		Security:    "reality",
		SNI:         "example.org",
		Fingerprint: "edge",
		PublicKey:   "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		ShortID:     "1a2b3c4d",
		NetworkType: "xhttp",
		Path:        "/path",
		Remark:      "TestServerCheck",
	}

	data, err := singbox.GenerateConfig(n, "tun0", []string{"1.1.1.1", "8.8.8.8"})
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "test-config.json")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	err = singbox.CheckConfig(tmpFile)
	if err != nil {
		t.Fatalf("singbox check failed: %v", err)
	}
}
