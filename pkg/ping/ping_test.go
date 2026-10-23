package ping_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"archvpn/pkg/node"
	"archvpn/pkg/ping"
)

func TestBenchmarkNodes(t *testing.T) {
	// Start a local mock TCP server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	goodNode := &node.Node{
		ID:     "good-node",
		Server: "127.0.0.1",
		Port:   port,
	}

	badNode := &node.Node{
		ID:     "bad-node",
		Server: "127.0.0.1",
		Port:   59999, // Unused port
	}

	resultsChan := ping.BenchmarkNodes([]*node.Node{goodNode, badNode}, 500*time.Millisecond, 2)

	var results []ping.BenchmarkResult
	for res := range resultsChan {
		results = append(results, res)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, res := range results {
		if res.Node.ID == "good-node" {
			if res.LatencyMs < 0 {
				t.Errorf("expected positive latency for good-node, got %d", res.LatencyMs)
			}
		} else if res.Node.ID == "bad-node" {
			if res.LatencyMs != -1 {
				t.Errorf("expected -1 latency for bad-node, got %d", res.LatencyMs)
			}
		}
	}
}
