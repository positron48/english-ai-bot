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

// telegramLongPollHTTPTimeouts returns timeouts for Telegram Bot API long polling (getUpdates).
// The server may not send response headers until the long-poll window ends or an update arrives,
// so ResponseHeaderTimeout and Client.Timeout must exceed the configured poll period.
func telegramLongPollHTTPTimeouts(updatesTimeoutSec int) (responseHeader time.Duration, clientTimeout time.Duration) {
	pollSec := updatesTimeoutSec
	if pollSec <= 0 {
		pollSec = 30
	}
	// https://core.telegram.org/bots/api#getupdates — timeout is optional, typical range 1–50s.
	if pollSec > 50 {
		pollSec = 50
	}
	const margin = 25 * time.Second
	responseHeader = time.Duration(pollSec)*time.Second + margin
	clientTimeout = responseHeader + 20*time.Second
	return responseHeader, clientTimeout
}

func newTelegramHTTPClientWithSocks5Proxy(proxyAddr string, updatesTimeoutSec int, log *zap.Logger) (*http.Client, error) {
	if proxyAddr == "" {
		return &http.Client{}, nil
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("init socks5 dialer for Telegram: %w", err)
	}

	log.Info("Telegram SOCKS5 proxy enabled", zap.String("addr", proxyAddr))

	headerTO, clientTO := telegramLongPollHTTPTimeouts(updatesTimeoutSec)

	transport := &http.Transport{
		Proxy: nil, // ensure we don't try to use proxy env vars in addition
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: headerTO,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          10,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   clientTO,
	}, nil
}
