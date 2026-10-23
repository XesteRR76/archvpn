package singbox

import (
	"encoding/json"
	"fmt"
	"strings"

	"archvpn/pkg/node"
)

// GenerateConfig generates a complete valid sing-box 1.14 JSON configuration
func GenerateConfig(n *node.Node, tunInterface string, dnsServers []string) ([]byte, error) {
	if n == nil {
		return nil, fmt.Errorf("node is nil")
	}
	if tunInterface == "" {
		tunInterface = "tun0"
	}

	remoteDNSHost := "1.1.1.1"
	if len(dnsServers) > 0 && dnsServers[0] != "" {
		remoteDNSHost = dnsServers[0]
	}

	// 1. Inbounds: TUN device
	inbounds := []map[string]interface{}{
		{
			"type":           "tun",
			"tag":            "tun-in",
			"interface_name": tunInterface,
			"address": []string{
				"172.19.0.1/30",
				"fdfe:dcba:9876::1/126",
			},
			"auto_route":   true,
			"strict_route": true,
		},
	}

	// 2. Outbounds: Proxy, Direct
	proxyOutbound := map[string]interface{}{
		"type":            n.Protocol,
		"tag":             "proxy",
		"server":          n.Server,
		"server_port":     n.Port,
		"uuid":            n.UUID,
		"packet_encoding": "xudp",
	}

	if n.Flow != "" {
		proxyOutbound["flow"] = n.Flow
	}

	// TLS / Reality setup
	if n.Security == "reality" || n.Security == "tls" {
		tlsConfig := map[string]interface{}{
			"enabled": true,
		}
		if n.SNI != "" {
			tlsConfig["server_name"] = n.SNI
		}
		if n.Fingerprint != "" {
			tlsConfig["utls"] = map[string]interface{}{
				"enabled":     true,
				"fingerprint": n.Fingerprint,
			}
		} else {
			tlsConfig["utls"] = map[string]interface{}{
				"enabled":     true,
				"fingerprint": "chrome",
			}
		}

		if n.Security == "reality" {
			realityConfig := map[string]interface{}{
				"enabled":    true,
				"public_key": n.PublicKey,
			}
			if n.ShortID != "" {
				realityConfig["short_id"] = n.ShortID
			}
			tlsConfig["reality"] = realityConfig
		}

		proxyOutbound["tls"] = tlsConfig
	}

	// Transports (xhttp/httpupgrade, ws, grpc)
	netType := strings.ToLower(n.NetworkType)
	switch netType {
	case "xhttp", "httpupgrade":
		proxyOutbound["transport"] = map[string]interface{}{
			"type": "httpupgrade",
			"path": n.Path,
			"host": n.Host,
		}
	case "ws":
		ws := map[string]interface{}{
			"type": "ws",
			"path": n.Path,
		}
		if n.Host != "" {
			ws["headers"] = map[string]interface{}{
				"Host": n.Host,
			}
		}
		proxyOutbound["transport"] = ws
	case "grpc":
		proxyOutbound["transport"] = map[string]interface{}{
			"type":         "grpc",
			"service_name": n.Path,
		}
	}

	outbounds := []map[string]interface{}{
		proxyOutbound,
		{
			"type": "direct",
			"tag":  "direct",
		},
	}

	// 3. DNS Configuration (sing-box 1.14 syntax)
	dnsBlock := map[string]interface{}{
		"servers": []map[string]interface{}{
			{
				"tag":    "dns-remote",
				"type":   "tls",
				"server": remoteDNSHost,
				"detour": "proxy",
			},
			{
				"tag":    "dns-local",
				"type":   "local",
				"detour": "direct",
			},
		},
		"strategy": "prefer_ipv4",
	}

	// 4. Routing block (with loop prevention for proxy server IP/domain)
	serverRule := map[string]interface{}{
		"outbound": "direct",
	}
	if strings.Contains(n.Server, ":") || strings.Count(n.Server, ".") == 3 {
		serverRule["ip_cidr"] = []string{n.Server + "/32"}
	} else {
		serverRule["domain"] = []string{n.Server}
	}

	routeRules := []map[string]interface{}{
		{
			"action": "sniff",
		},
		{
			"protocol": "dns",
			"action":   "hijack-dns",
		},
		serverRule,
		{
			"ip_is_private": true,
			"outbound":      "direct",
		},
	}

	routeBlock := map[string]interface{}{
		"auto_detect_interface":   true,
		"default_domain_resolver": "dns-local",
		"rules":                   routeRules,
	}

	fullConfig := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"dns":       dnsBlock,
		"inbounds":  inbounds,
		"outbounds": outbounds,
		"route":     routeBlock,
	}

	return json.MarshalIndent(fullConfig, "", "  ")
}
