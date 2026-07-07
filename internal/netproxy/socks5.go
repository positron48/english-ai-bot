package netproxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const DefaultRetryAttempts = 3

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
		DisableKeepAlives:     true,
	}, nil
}

func DoJSONWithRetry(ctx context.Context, client *http.Client, method, endpoint string, body []byte, headers map[string]string) (*http.Response, error) {
	if client == nil {
		client = http.DefaultClient
	}
	attempts := DefaultRetryAttempts
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if ctx.Err() != nil || attempt == attempts || !isRetryableTransportError(err) {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt*250) * time.Millisecond):
		}
	}
	return nil, lastErr
}

func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return isRetryableTransportError(urlErr.Err)
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "unexpected eof")
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
