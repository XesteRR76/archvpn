package node

import (
	"fmt"
	"strings"
)

// Node represents a parsed proxy server endpoint
type Node struct {
	ID          string `json:"id"`
	Protocol    string `json:"protocol"`
	UUID        string `json:"uuid"`
	Server      string `json:"server"`
	Port        int    `json:"port"`
	Encryption  string `json:"encryption"`
	Flow        string `json:"flow"`
	Security    string `json:"security"`
	SNI         string `json:"sni"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
	ShortID     string `json:"short_id"`
	SpiderX     string `json:"spider_x"`
	NetworkType string `json:"network_type"`
	Path        string `json:"path"`
	Host        string `json:"host"`
	Remark      string `json:"remark"`
	RawURL      string `json:"raw_url"`
	LatencyMs   int64  `json:"latency_ms"` // -1 for timeout / unreachable
}

// DisplayName returns a formatted human-readable name with flag/prefix if available
func (n *Node) DisplayName() string {
	if n.Remark != "" {
		return n.Remark
	}
	return fmt.Sprintf("%s:%d", n.Server, n.Port)
}

// InferCountry attempts to identify region/country tag from the remark
func (n *Node) InferCountry() string {
	r := strings.ToUpper(n.Remark)

	// Check if remark has flag emojis or Russian country names
	switch {
	case strings.Contains(r, "🇳🇱") || strings.Contains(r, "НИДЕРЛАНД") || strings.Contains(r, "NETHERLAND") || strings.Contains(r, "AMSTERDAM"):
		return "🇳🇱 NL"
	case strings.Contains(r, "🇩🇪") || strings.Contains(r, "ГЕРМАНИ") || strings.Contains(r, "GERMAN") || strings.Contains(r, "ФРАНКФУРТ") || strings.Contains(r, "БАВАРИЯ"):
		return "🇩🇪 DE"
	case strings.Contains(r, "🇷🇺") || strings.Contains(r, "РОССИ") || strings.Contains(r, "RUSSIA") || strings.Contains(r, "МОСКВ") || strings.Contains(r, "СПБ"):
		return "🇷🇺 RU"
	case strings.Contains(r, "🇫🇮") || strings.Contains(r, "ФИНЛЯНД") || strings.Contains(r, "FINLAND") || strings.Contains(r, "ХЕЛЬСИНКИ"):
		return "🇫🇮 FI"
	case strings.Contains(r, "🇪🇪") || strings.Contains(r, "ЭСТОНИ") || strings.Contains(r, "ESTONIA") || strings.Contains(r, "ТАЛЛИН"):
		return "🇪🇪 EE"
	case strings.Contains(r, "🇵🇱") || strings.Contains(r, "ПОЛЬШ") || strings.Contains(r, "POLAND") || strings.Contains(r, "ВАРШАВ"):
		return "🇵🇱 PL"
	case strings.Contains(r, "🇨🇭") || strings.Contains(r, "ШВЕЙЦАР") || strings.Contains(r, "SWISS") || strings.Contains(r, "SWITZERLAND"):
		return "🇨🇭 CH"
	case strings.Contains(r, "🇸🇪") || strings.Contains(r, "ШВЕЦИ") || strings.Contains(r, "SWEDEN"):
		return "🇸🇪 SE"
	case strings.Contains(r, "🇺🇸") || strings.Contains(r, "США") || strings.Contains(r, "UNITED STATES") || strings.Contains(r, "AMERICA") || strings.Contains(r, "ДАЛЛАС"):
		return "🇺🇸 US"
	case strings.Contains(r, "🇧🇾") || strings.Contains(r, "БЕЛАРУС") || strings.Contains(r, "BELARUS") || strings.Contains(r, "МИНСК"):
		return "🇧🇾 BY"
	case strings.Contains(r, "🇱🇹") || strings.Contains(r, "ЛИТВ") || strings.Contains(r, "LITHUANIA"):
		return "🇱🇹 LT"
	case strings.Contains(r, "🇬🇧") || strings.Contains(r, "UK") || strings.Contains(r, "LONDON"):
		return "🇬🇧 UK"
	case strings.Contains(r, "🇹🇷") || strings.Contains(r, "ТУРЦИ") || strings.Contains(r, "TURKEY"):
		return "🇹🇷 TR"
	case strings.Contains(r, "🇰🇿") || strings.Contains(r, "КАЗАХСТАН") || strings.Contains(r, "KAZAKHSTAN"):
		return "🇰🇿 KZ"
	case strings.Contains(r, "🇸🇬") || strings.Contains(r, "СИНГАПУР") || strings.Contains(r, "SINGAPORE"):
		return "🇸🇬 SG"
	case strings.Contains(r, "🇯🇵") || strings.Contains(r, "ЯПОНИ") || strings.Contains(r, "JAPAN"):
		return "🇯🇵 JP"
	case strings.Contains(r, "🇪🇺") || strings.Contains(r, "ЕВРОП") || strings.Contains(r, "EUROPE"):
		return "🇪🇺 EU"
	default:
		return "🌐 Global"
	}
}
