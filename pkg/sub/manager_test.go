package sub_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"archvpn/pkg/config"
	"archvpn/pkg/sub"
)

func TestSubscriptionManager(t *testing.T) {
	nodeStr := "vless://00000000-0000-0000-0000-000000000001@192.0.2.1:443?security=reality&pbk=key1#MockServer"
	b64Content := base64.StdEncoding.EncodeToString([]byte(nodeStr))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(b64Content))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		ConfigDir:        tmpDir,
		SubscriptionURLs: []string{ts.URL},
	}

	mgr := sub.NewManager(cfg)

	// Test Update
	err := mgr.Update()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	nodes := mgr.GetNodes()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Remark != "MockServer" {
		t.Errorf("expected remark MockServer, got %s", nodes[0].Remark)
	}

	// Verify nodes were cached to disk
	cacheFile := filepath.Join(tmpDir, "nodes.json")
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Fatalf("nodes.json was not written to disk")
	}

	// Test reload from cache
	mgr2 := sub.NewManager(cfg)
	if len(mgr2.GetNodes()) != 1 {
		t.Errorf("expected 1 loaded node from cache, got %d", len(mgr2.GetNodes()))
	}

	// Test FindNode by 1-based index
	found := mgr2.FindNode("1")
	if found == nil || found.Remark != "MockServer" {
		t.Errorf("expected to find node by index 1, got %v", found)
	}
}
