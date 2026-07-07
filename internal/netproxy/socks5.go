package netproxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// NewHTTPClient returns an HTTP client with an optional SOCKS5 proxy transport.
// Expected proxy formats: "ip:port", "socks5://ip:port", "socks5://user:pass@ip:port".
func NewHTTPClient(timeout time.Duration, socks5Proxy string) (*http.Client, error) {
	client := &http.Client{Timeout: timeout}
	if strings.TrimSpace(socks5Proxy) == "" {
		return client, nil
	}
	transport, err := NewSocks5Transport(socks5Proxy)
	if err != nil {
		return client, err
	}
	client.Transport = transport
	return client, nil
}

func NewSocks5Transport(raw string) (*http.Transport, error) {
	hostPort, auth, err := normalizeSocks5Proxy(raw)
	if err != nil {
		return nil, err
	}
	dialer, err := proxy.SOCKS5("tcp", hostPort, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &http.Transport{
		Proxy: nil,
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          10,
	}, nil
}

func normalizeSocks5Proxy(raw string) (hostPort string, auth *proxy.Auth, err error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil, fmt.Errorf("empty socks5 proxy")
	}
	if !strings.Contains(value, "://") {
		return value, nil, nil
	}
	u, err := url.Parse(value)
	if err != nil {
		return "", nil, fmt.Errorf("parse socks5 proxy: %w", err)
	}
	if u.Scheme != "socks5" && u.Scheme != "socks5h" {
		return "", nil, fmt.Errorf("unsupported socks5 proxy scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return "", nil, fmt.Errorf("socks5 proxy host is empty")
	}
	if u.User != nil {
		password, _ := u.User.Password()
		auth = &proxy.Auth{User: u.User.Username(), Password: password}
	}
	return u.Host, auth, nil
}
