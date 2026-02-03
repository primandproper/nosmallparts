package tracing

import (
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	defaultTimeout = 30 * time.Second
)

// BuildTracedHTTPTransport constructs a new http.Transport.
func BuildTracedHTTPTransport(timeout time.Duration) http.RoundTripper {
	if timeout == 0 {
		timeout = defaultTimeout
	}

	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 2 * timeout,
		IdleConnTimeout:       3 * timeout,
	}

	return otelhttp.NewTransport(t, otelhttp.WithSpanNameFormatter(FormatSpan))
}

// BuildTracedHTTPClient returns a tracing-enabled HTTP client.
func BuildTracedHTTPClient() *http.Client {
	return &http.Client{Transport: BuildTracedHTTPTransport(10 * time.Second), Timeout: 10 * time.Second}
}
