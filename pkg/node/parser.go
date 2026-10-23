package node

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ParseSubscription decodes raw subscription content (base64 or plain lines) and parses all nodes
func ParseSubscription(content string) ([]*Node, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, fmt.Errorf("empty subscription content")
	}

	decodedText := tryBase64Decode(trimmed)
	lines := strings.Split(strings.ReplaceAll(decodedText, "\r\n", "\n"), "\n")

	var nodes []*Node
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "vless://") {
			n, err := ParseVLESS(line)
			if err == nil && n != nil {
				if !seen[n.ID] {
					seen[n.ID] = true
					nodes = append(nodes, n)
				}
			}
		}
	}

	return nodes, nil
}

// ParseVLESS parses a standard vless:// URI into a Node
func ParseVLESS(rawURL string) (*Node, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "vless" {
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	if u.User == nil {
		return nil, fmt.Errorf("missing userinfo / uuid in URL")
	}

	uuid := u.User.Username()
	if uuid == "" {
		return nil, fmt.Errorf("empty uuid in URL")
	}
	host := u.Hostname()
	portStr := u.Port()

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", portStr, err)
	}

	q := u.Query()

	remark := u.Fragment
	if unescaped, err := url.QueryUnescape(remark); err == nil {
		remark = unescaped
	}

	// Generate stable deterministic ID
	hashInput := fmt.Sprintf("%s:%s:%d:%s", uuid, host, port, remark)
	h := sha256.Sum256([]byte(hashInput))
	nodeID := hex.EncodeToString(h[:8])

	n := &Node{
		ID:          nodeID,
		Protocol:    "vless",
		UUID:        uuid,
		Server:      host,
		Port:        port,
		Encryption:  q.Get("encryption"),
		Flow:        q.Get("flow"),
		Security:    q.Get("security"),
		SNI:         q.Get("sni"),
		Fingerprint: q.Get("fp"),
		PublicKey:   q.Get("pbk"),
		ShortID:     q.Get("sid"),
		SpiderX:     q.Get("spx"),
		NetworkType: q.Get("type"),
		Path:        q.Get("path"),
		Host:        q.Get("host"),
		Remark:      remark,
		RawURL:      rawURL,
		LatencyMs:   -1,
	}

	if n.NetworkType == "" {
		n.NetworkType = "tcp"
	}
	if n.Encryption == "" {
		n.Encryption = "none"
	}

	return n, nil
}

// tryBase64Decode tries standard, URL-safe, and unpadded base64 decoders
func tryBase64Decode(input string) string {
	clean := strings.ReplaceAll(strings.ReplaceAll(input, "\n", ""), "\r", "")
	clean = strings.TrimSpace(clean)

	// If already starts with a protocol scheme, return as is
	if strings.HasPrefix(clean, "vless://") || strings.HasPrefix(clean, "vmess://") {
		return input
	}

	// Pad if needed
	pad := len(clean) % 4
	if pad > 0 {
		clean += strings.Repeat("=", 4-pad)
	}

	// Try StdEncoding
	if b, err := base64.StdEncoding.DecodeString(clean); err == nil {
		return string(b)
	}

	// Try URLEncoding
	if b, err := base64.URLEncoding.DecodeString(clean); err == nil {
		return string(b)
	}

	// Try RawStdEncoding
	cleanTrimmed := strings.TrimRight(clean, "=")
	if b, err := base64.RawStdEncoding.DecodeString(cleanTrimmed); err == nil {
		return string(b)
	}

	// Try RawURLEncoding
	if b, err := base64.RawURLEncoding.DecodeString(cleanTrimmed); err == nil {
		return string(b)
	}

	return input
}
