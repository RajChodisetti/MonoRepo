package seoreport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const websiteDNSBudget = 1200 * time.Millisecond

type websiteHostResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type websiteContextDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

var nonPublicInternetRanges = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001:db8::/32"),
}

// validatePublicWebsiteURL rejects every destination that a public report is
// not allowed to reach. DNS names are accepted only when every answer is a
// public unicast address; mixed public/private answers therefore fail closed.
func validatePublicWebsiteURL(ctx context.Context, rawURL string, resolver websiteHostResolver) (*url.URL, error) {
	if resolver == nil {
		return nil, errors.New("website resolver unavailable")
	}
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" || strings.ContainsAny(rawURL, "\x00\r\n") {
		return nil, errors.New("invalid website URL")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil || parsed.Opaque != "" {
		return nil, errors.New("invalid website URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("website URL must use http or https")
	}
	if parsed.User != nil {
		return nil, errors.New("website URL credentials are not allowed")
	}
	host := parsed.Hostname()
	if host == "" {
		return nil, errors.New("website URL host is required")
	}
	if port := parsed.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return nil, errors.New("website URL port is invalid")
		}
	}
	if _, err := resolvePublicWebsiteHost(ctx, host, resolver); err != nil {
		return nil, err
	}
	return parsed, nil
}

func resolvePublicWebsiteHost(ctx context.Context, rawHost string, resolver websiteHostResolver) ([]netip.Addr, error) {
	host := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(rawHost), "."))
	if host == "" || strings.ContainsAny(host, "\x00\r\n/%") {
		return nil, errors.New("invalid website host")
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "localhost.localdomain" {
		return nil, errors.New("local website host is not allowed")
	}

	if literal, err := netip.ParseAddr(host); err == nil {
		literal = literal.Unmap()
		if !isPublicInternetAddress(literal) {
			return nil, errors.New("non-public website address is not allowed")
		}
		return []netip.Addr{literal}, nil
	}
	// Single-label names can be resolved through container/search domains (for
	// example postgres, api, or metadata) and are never valid public websites.
	if !strings.Contains(host, ".") {
		return nil, errors.New("single-label website host is not allowed")
	}

	answers, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve website host: %w", err)
	}
	if len(answers) == 0 {
		return nil, errors.New("website host has no addresses")
	}
	resolved := make([]netip.Addr, 0, len(answers))
	seen := make(map[netip.Addr]struct{}, len(answers))
	for _, answer := range answers {
		if answer.Zone != "" {
			return nil, errors.New("scoped website address is not allowed")
		}
		addr, ok := netip.AddrFromSlice(answer.IP)
		if !ok {
			return nil, errors.New("website host returned an invalid address")
		}
		addr = addr.Unmap()
		if !isPublicInternetAddress(addr) {
			return nil, errors.New("website host resolved to a non-public address")
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		resolved = append(resolved, addr)
	}
	return resolved, nil
}

func isPublicInternetAddress(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	for _, blocked := range nonPublicInternetRanges {
		if blocked.Contains(addr) {
			return false
		}
	}
	return true
}

// guardBrowserRequest is used by Fetch.requestPaused handling. The callback is
// intentionally injectable so redirect/subresource enforcement is testable
// without launching a browser.
func guardBrowserRequest(
	ctx context.Context,
	resolver websiteHostResolver,
	rawURL string,
	allow func(context.Context) error,
	block func(context.Context) error,
) error {
	validationCtx, cancel := context.WithTimeout(ctx, websiteDNSBudget)
	_, err := validatePublicWebsiteURL(validationCtx, rawURL, resolver)
	cancel()
	if err != nil {
		return block(ctx)
	}
	return allow(ctx)
}

// safeBrowserProxy is the network enforcement boundary for Chromium. It
// resolves each destination itself, validates every answer, and dials a vetted
// IP literal. Chromium never connects to a hostname directly, closing the DNS
// rebinding gap between request interception and the actual socket connect.
type safeBrowserProxy struct {
	resolver  websiteHostResolver
	dialer    websiteContextDialer
	server    *http.Server
	listener  net.Listener
	transport *http.Transport

	mu      sync.Mutex
	tunnels map[net.Conn]struct{}
}

func startSafeBrowserProxy(resolver websiteHostResolver) (*safeBrowserProxy, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	proxy := newSafeBrowserProxy(resolver, &net.Dialer{
		Timeout:   2500 * time.Millisecond,
		KeepAlive: 15 * time.Second,
	})
	proxy.listener = listener
	proxy.server = &http.Server{
		Handler:           proxy,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       10 * time.Second,
	}
	go func() {
		_ = proxy.server.Serve(listener)
	}()
	return proxy, nil
}

func newSafeBrowserProxy(resolver websiteHostResolver, dialer websiteContextDialer) *safeBrowserProxy {
	proxy := &safeBrowserProxy{
		resolver: resolver,
		dialer:   dialer,
		tunnels:  make(map[net.Conn]struct{}),
	}
	proxy.transport = &http.Transport{
		Proxy:                 nil,
		DialContext:           proxy.dialContext,
		ForceAttemptHTTP2:     false,
		DisableCompression:    false,
		TLSHandshakeTimeout:   2500 * time.Millisecond,
		ResponseHeaderTimeout: 4 * time.Second,
		IdleConnTimeout:       10 * time.Second,
	}
	return proxy
}

func (proxy *safeBrowserProxy) URL() string {
	if proxy == nil || proxy.listener == nil {
		return ""
	}
	return "http://" + proxy.listener.Addr().String()
}

func (proxy *safeBrowserProxy) Close() {
	if proxy == nil {
		return
	}
	if proxy.transport != nil {
		proxy.transport.CloseIdleConnections()
	}
	if proxy.server != nil {
		_ = proxy.server.Close()
	}
	proxy.mu.Lock()
	for conn := range proxy.tunnels {
		_ = conn.Close()
	}
	proxy.tunnels = make(map[net.Conn]struct{})
	proxy.mu.Unlock()
}

func (proxy *safeBrowserProxy) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodConnect {
		proxy.serveConnect(w, request)
		return
	}
	if request.URL == nil || !request.URL.IsAbs() {
		http.Error(w, "blocked destination", http.StatusForbidden)
		return
	}
	validationCtx, cancel := context.WithTimeout(request.Context(), websiteDNSBudget)
	_, err := validatePublicWebsiteURL(validationCtx, request.URL.String(), proxy.resolver)
	cancel()
	if err != nil {
		http.Error(w, "blocked destination", http.StatusForbidden)
		return
	}

	outbound := request.Clone(request.Context())
	outbound.RequestURI = ""
	outbound.Host = outbound.URL.Host
	removeProxyHopHeaders(outbound.Header)
	response, err := proxy.transport.RoundTrip(outbound)
	if err != nil {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	copyHTTPHeader(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

func (proxy *safeBrowserProxy) serveConnect(w http.ResponseWriter, request *http.Request) {
	host, port, err := net.SplitHostPort(request.Host)
	if err != nil || host == "" || port == "" {
		http.Error(w, "blocked destination", http.StatusForbidden)
		return
	}
	upstream, err := proxy.dialPublicHost(request.Context(), "tcp", host, port)
	if err != nil {
		http.Error(w, "blocked destination", http.StatusForbidden)
		return
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		_ = upstream.Close()
		http.Error(w, "proxy tunnel unavailable", http.StatusInternalServerError)
		return
	}
	client, buffered, err := hijacker.Hijack()
	if err != nil {
		_ = upstream.Close()
		return
	}
	proxy.trackTunnel(client, true)
	proxy.trackTunnel(upstream, true)
	defer func() {
		proxy.trackTunnel(client, false)
		proxy.trackTunnel(upstream, false)
		_ = client.Close()
		_ = upstream.Close()
	}()
	if _, err := buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		return
	}
	if err := buffered.Flush(); err != nil {
		return
	}

	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(upstream, buffered)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(client, upstream)
		done <- struct{}{}
	}()
	<-done
}

func (proxy *safeBrowserProxy) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	return proxy.dialPublicHost(ctx, network, host, port)
}

func (proxy *safeBrowserProxy) dialPublicHost(ctx context.Context, network, host, port string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, errors.New("unsupported website network")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return nil, errors.New("invalid website port")
	}
	resolveCtx, cancel := context.WithTimeout(ctx, websiteDNSBudget)
	addresses, err := resolvePublicWebsiteHost(resolveCtx, host, proxy.resolver)
	cancel()
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, address := range addresses {
		conn, dialErr := proxy.dialer.DialContext(ctx, network, net.JoinHostPort(address.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr == nil {
		lastErr = errors.New("website address unavailable")
	}
	return nil, lastErr
}

func (proxy *safeBrowserProxy) trackTunnel(conn net.Conn, add bool) {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	if add {
		proxy.tunnels[conn] = struct{}{}
	} else {
		delete(proxy.tunnels, conn)
	}
}

func removeProxyHopHeaders(header http.Header) {
	for _, value := range header.Values("Connection") {
		for _, name := range strings.Split(value, ",") {
			header.Del(strings.TrimSpace(name))
		}
	}
	for _, name := range []string{
		"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate",
		"Proxy-Authorization", "TE", "Trailer", "Transfer-Encoding", "Upgrade",
	} {
		header.Del(name)
	}
}

func copyHTTPHeader(destination, source http.Header) {
	for key, values := range source {
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}
