package sub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"archvpn/pkg/config"
	"archvpn/pkg/node"
)

type Manager struct {
	cfg        *config.Config
	nodes      []*node.Node
	httpClient *http.Client
}

func NewManager(cfg *config.Config) *Manager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	m := &Manager{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
	_ = m.loadCache()
	return m
}

// GetNodes returns the current cached nodes
func (m *Manager) GetNodes() []*node.Node {
	return m.nodes
}

// Update downloads all subscriptions and refreshes the node cache
func (m *Manager) Update() error {
	var allNodes []*node.Node
	seen := make(map[string]bool)
	var errs []string

	for _, subURL := range m.cfg.SubscriptionURLs {
		if strings.TrimSpace(subURL) == "" {
			continue
		}
		nodes, err := m.fetchSubscription(subURL)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", subURL, err))
			continue
		}

		for _, n := range nodes {
			if !seen[n.ID] {
				seen[n.ID] = true
				allNodes = append(allNodes, n)
			}
		}
	}

	if len(allNodes) == 0 && len(errs) > 0 {
		return fmt.Errorf("failed to fetch subscriptions:\n%s", strings.Join(errs, "\n"))
	}

	m.nodes = allNodes
	_ = m.SaveNodes()
	return nil
}

// SaveNodes saves nodes to nodes.json
func (m *Manager) SaveNodes() error {
	if m.cfg.ConfigDir == "" {
		return nil
	}
	_ = os.MkdirAll(m.cfg.ConfigDir, 0755)
	cacheFile := filepath.Join(m.cfg.ConfigDir, "nodes.json")

	data, err := json.MarshalIndent(m.nodes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacheFile, data, 0644)
}

func (m *Manager) loadCache() error {
	if m.cfg.ConfigDir == "" {
		return nil
	}
	cacheFile := filepath.Join(m.cfg.ConfigDir, "nodes.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.nodes)
}

func (m *Manager) fetchSubscription(url string) ([]*node.Node, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Use standard client UA to ensure servers don't reject
	req.Header.Set("User-Agent", "sing-box/1.14.3")
	req.Header.Set("Accept", "*/*")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return node.ParseSubscription(string(body))
}

// FindNode finds a node by 1-based index, ID or remark
func (m *Manager) FindNode(query string) *node.Node {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil
	}

	// 1. Try 1-based numeric index
	if idx, err := strconv.Atoi(trimmed); err == nil {
		if idx >= 1 && idx <= len(m.nodes) {
			return m.nodes[idx-1]
		}
	}

	for _, n := range m.nodes {
		if strings.EqualFold(n.ID, trimmed) || strings.HasPrefix(strings.ToLower(n.ID), strings.ToLower(trimmed)) {
			return n
		}
		if strings.Contains(strings.ToLower(n.Remark), strings.ToLower(trimmed)) {
			return n
		}
	}
	return nil
}
