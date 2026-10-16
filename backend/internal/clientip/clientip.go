// Package clientip says which address a request came from, for bans and rate
// limits.
//
// Behind a reverse proxy the socket address is the proxy's, so the last entry
// of X-Forwarded-For is trusted instead: it is the hop the proxy closest to us
// appended, which a client cannot forge. Without the header the socket address
// is used. If a proxy in front stops sending it, everybody shares the proxy's
// address, and a ban or a rate limit lands on everyone at once; that is the
// loud failure, preferred to trusting a header a client wrote.
package clientip

import (
	"context"
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// forwardedFor is the header, and the gRPC metadata key it arrives as.
const forwardedFor = "x-forwarded-for"

// FromRequest is the client address of an HTTP request.
func FromRequest(r *http.Request) string {
	if ip := lastForwarded(r.Header.Values(forwardedFor)); ip != "" {
		return ip
	}

	return host(r.RemoteAddr)
}

// FromContext is the client address of a gRPC call: the forwarded header the
// call carried as metadata, or its peer.
func FromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ip := lastForwarded(md.Get(forwardedFor)); ip != "" {
			return ip
		}
	}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		return host(p.Addr.String())
	}

	return ""
}

// lastForwarded is the last non-empty entry across every X-Forwarded-For
// value, since a proxy may append a header rather than extend one.
func lastForwarded(values []string) string {
	last := ""

	for _, value := range values {
		for entry := range strings.SplitSeq(value, ",") {
			if trimmed := strings.TrimSpace(entry); trimmed != "" {
				last = trimmed
			}
		}
	}

	return last
}

func host(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}

	return addr
}
