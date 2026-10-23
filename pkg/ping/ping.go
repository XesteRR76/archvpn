package ping

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"archvpn/pkg/node"
)

type BenchmarkResult struct {
	Node      *node.Node
	LatencyMs int64
	Err       error
}

// PingNode tests TCP or TLS connection latency to a single node
func PingNode(n *node.Node, timeout time.Duration) int64 {
	addr := net.JoinHostPort(n.Server, fmt.Sprintf("%d", n.Port))
	start := time.Now()

	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		n.LatencyMs = -1
		return -1
	}

	// If reality/tls with SNI is specified, optionally measure TLS handshake
	if (n.Security == "reality" || n.Security == "tls") && n.SNI != "" {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName:         n.SNI,
			InsecureSkipVerify: true,
		})
		_ = tlsConn.SetDeadline(time.Now().Add(timeout))
		if err := tlsConn.Handshake(); err == nil {
			elapsed := time.Since(start).Milliseconds()
			_ = tlsConn.Close()
			n.LatencyMs = elapsed
			return elapsed
		}
	}

	elapsed := time.Since(start).Milliseconds()
	_ = conn.Close()
	n.LatencyMs = elapsed
	return elapsed
}

// BenchmarkNodes concurrently tests latency for a slice of nodes
func BenchmarkNodes(nodes []*node.Node, timeout time.Duration, concurrency int) <-chan BenchmarkResult {
	if concurrency <= 0 {
		concurrency = 20
	}
	if timeout <= 0 {
		timeout = 2 * time.Second
	}

	out := make(chan BenchmarkResult, len(nodes))
	jobs := make(chan *node.Node, len(nodes))

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range jobs {
				lat := PingNode(n, timeout)
				var err error
				if lat == -1 {
					err = fmt.Errorf("timeout / unreachable")
				}
				out <- BenchmarkResult{
					Node:      n,
					LatencyMs: lat,
					Err:       err,
				}
			}
		}()
	}

	for _, n := range nodes {
		jobs <- n
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
