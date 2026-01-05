package web

import (
	"net"
	"net/http"
	"strings"
)

// clientIP extracts the real client IP from the request, considering nginx proxy headers
func clientIP(r *http.Request) string {
	// Priority 1: X-Real-IP (nginx sets this directly)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if parsed := parseIP(ip); parsed != "" {
			return parsed
		}
	}

	// Priority 2: X-Forwarded-For (first IP in the chain)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsed := parseIP(ip); parsed != "" {
				return parsed
			}
		}
	}

	// Priority 3: RemoteAddr (fallback)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, try using RemoteAddr directly (might not have port)
		ip = r.RemoteAddr
	}
	if parsed := parseIP(ip); parsed != "" {
		return parsed
	}

	// Ultimate fallback
	return "unknown"
}

// parseIP validates and returns a clean IP address, or empty string if invalid
func parseIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}

	// Remove port if present
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	// Validate it's a valid IP
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}

	// Return IPv4 in string form, or IPv6
	return parsed.String()
}
