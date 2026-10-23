package node_test

import (
	"encoding/base64"
	"testing"

	"archvpn/pkg/node"
)

func TestParseVLESS(t *testing.T) {
	raw := "vless://00000000-0000-0000-0000-000000000001@192.0.2.1:443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=example.com&fp=chrome&pbk=testpubkey123&sid=1a2b3c&type=tcp#NL-Amsterdam"

	n, err := node.ParseVLESS(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n.Protocol != "vless" {
		t.Errorf("expected protocol vless, got %s", n.Protocol)
	}
	if n.UUID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("unexpected uuid: %s", n.UUID)
	}
	if n.Server != "192.0.2.1" {
		t.Errorf("unexpected server: %s", n.Server)
	}
	if n.Port != 443 {
		t.Errorf("unexpected port: %d", n.Port)
	}
	if n.Security != "reality" {
		t.Errorf("unexpected security: %s", n.Security)
	}
	if n.Flow != "xtls-rprx-vision" {
		t.Errorf("unexpected flow: %s", n.Flow)
	}
	if n.PublicKey != "testpubkey123" {
		t.Errorf("unexpected pbk: %s", n.PublicKey)
	}
	if n.ShortID != "1a2b3c" {
		t.Errorf("unexpected sid: %s", n.ShortID)
	}
	if n.SNI != "example.com" {
		t.Errorf("unexpected sni: %s", n.SNI)
	}
	if n.Fingerprint != "chrome" {
		t.Errorf("unexpected fp: %s", n.Fingerprint)
	}
	if n.Remark != "NL-Amsterdam" {
		t.Errorf("unexpected remark: %s", n.Remark)
	}
}

func TestParseVLESS_MalformedNoUser(t *testing.T) {
	raw := "vless://example.com:443?security=none"
	_, err := node.ParseVLESS(raw)
	if err == nil {
		t.Errorf("expected error for VLESS without userinfo, got nil")
	}
}

func TestParseSubscription(t *testing.T) {
	raw1 := "vless://00000000-0000-0000-0000-000000000001@192.0.2.1:443?security=reality&pbk=k1#Node1"
	raw2 := "vless://00000000-0000-0000-0000-000000000002@192.0.2.2:8443?security=tls&sni=example.com#Node2"
	lines := raw1 + "\n" + raw2

	encoded := base64.StdEncoding.EncodeToString([]byte(lines))
	nodes, err := node.ParseSubscription(encoded)
	if err != nil {
		t.Fatalf("ParseSubscription failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Remark != "Node1" {
		t.Errorf("expected Node1, got %s", nodes[0].Remark)
	}
	if nodes[1].Remark != "Node2" {
		t.Errorf("expected Node2, got %s", nodes[1].Remark)
	}
}
