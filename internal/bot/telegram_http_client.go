package bot

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

func newTelegramHTTPClientWithSocks5Proxy(proxyAddr string, log *zap.Logger) (*http.Client, error) {
	if proxyAddr == "" {
		return &http.Client{}, nil
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("init socks5 dialer for Telegram: %w", err)
	}

	log.Info("Telegram SOCKS5 proxy enabled", zap.String("addr", proxyAddr))

	transport := &http.Transport{
		Proxy: nil, // ensure we don't try to use proxy env vars in addition
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		// Reasonable timeouts to avoid goroutine leaks when proxy is misconfigured.
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          10,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

