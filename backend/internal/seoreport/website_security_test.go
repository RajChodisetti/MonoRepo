package seoreport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type staticWebsiteResolver struct {
	answers map[string][]net.IPAddr
	err     error
}

func (resolver staticWebsiteResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	if resolver.err != nil {
		return nil, resolver.err
	}
	return resolver.answers[host], nil
}

type sequenceWebsiteResolver struct {
	mu      sync.Mutex
	answers [][]net.IPAddr
	calls   int
}

func (resolver *sequenceWebsiteResolver) LookupIPAddr(_ context.Context, _ string) ([]net.IPAddr, error) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	if resolver.calls >= len(resolver.answers) {
		return nil, errors.New("unexpected DNS lookup")
	}
	answer := resolver.answers[resolver.calls]
	resolver.calls++
	return answer, nil
}

type recordingWebsiteDialer struct {
	mu      sync.Mutex
	calls   int
	address string
	err     error
}

func (dialer *recordingWebsiteDialer) DialContext(_ context.Context, _, address string) (net.Conn, error) {
	dialer.mu.Lock()
	defer dialer.mu.Unlock()
	dialer.calls++
	dialer.address = address
	return nil, dialer.err
}

func TestValidatePublicWebsiteURLRejectsUnsafeDestinations(t *testing.T) {
	publicResolver := staticWebsiteResolver{answers: map[string][]net.IPAddr{
		"public.example": {{IP: net.ParseIP("93.184.216.34")}},
	}}
	tests := []struct {
		name string
		url  string
	}{
		{name: "non HTTP scheme", url: "file:///etc/passwd"},
		{name: "credentials", url: "https://user:pass@8.8.8.8/"},
		{name: "localhost", url: "http://localhost/admin"},
		{name: "localhost subdomain", url: "http://app.localhost/admin"},
		{name: "compose hostname", url: "http://postgres:5432/"},
		{name: "loopback IPv4", url: "http://127.0.0.1/"},
		{name: "private IPv4", url: "http://10.0.0.4/"},
		{name: "metadata IPv4", url: "http://169.254.169.254/latest/meta-data/"},
		{name: "unspecified IPv4", url: "http://0.0.0.0/"},
		{name: "loopback IPv6", url: "http://[::1]/"},
		{name: "private IPv6", url: "http://[fd00:ec2::254]/"},
		{name: "link local IPv6", url: "http://[fe80::1]/"},
		{name: "mapped loopback IPv6", url: "http://[::ffff:127.0.0.1]/"},
		{name: "multicast IPv6", url: "http://[ff02::1]/"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := validatePublicWebsiteURL(context.Background(), test.url, publicResolver); err == nil {
				t.Fatalf("validatePublicWebsiteURL(%q) succeeded, want rejection", test.url)
			}
		})
	}
	if _, err := validatePublicWebsiteURL(context.Background(), "https://8.8.8.8/", publicResolver); err != nil {
		t.Fatalf("public literal rejected: %v", err)
	}
	if _, err := validatePublicWebsiteURL(context.Background(), "https://public.example/menu", publicResolver); err != nil {
		t.Fatalf("public DNS destination rejected: %v", err)
	}
}

func TestValidatePublicWebsiteURLRejectsPrivateAndMixedDNSAnswers(t *testing.T) {
	resolver := staticWebsiteResolver{answers: map[string][]net.IPAddr{
		"private.example": {{IP: net.ParseIP("192.168.1.10")}},
		"mixed.example": {
			{IP: net.ParseIP("93.184.216.34")},
			{IP: net.ParseIP("127.0.0.1")},
		},
	}}
	for _, rawURL := range []string{"https://private.example", "https://mixed.example"} {
		if _, err := validatePublicWebsiteURL(context.Background(), rawURL, resolver); err == nil {
			t.Fatalf("validatePublicWebsiteURL(%q) accepted unsafe DNS answers", rawURL)
		}
	}
}

func TestGuardBrowserRequestBlocksPrivateRedirectAndSubresource(t *testing.T) {
	resolver := staticWebsiteResolver{answers: map[string][]net.IPAddr{
		"restaurant.example": {{IP: net.ParseIP("93.184.216.34")}},
	}}
	allowed, blocked := 0, 0
	allow := func(context.Context) error {
		allowed++
		return nil
	}
	block := func(context.Context) error {
		blocked++
		return nil
	}

	if err := guardBrowserRequest(context.Background(), resolver, "https://restaurant.example/", allow, block); err != nil {
		t.Fatalf("allow public navigation: %v", err)
	}
	// A redirect hop and a page subresource are independently intercepted.
	if err := guardBrowserRequest(context.Background(), resolver, "http://169.254.169.254/latest/meta-data/", allow, block); err != nil {
		t.Fatalf("block redirect: %v", err)
	}
	if err := guardBrowserRequest(context.Background(), resolver, "http://localhost:8080/internal", allow, block); err != nil {
		t.Fatalf("block subresource: %v", err)
	}
	if allowed != 1 || blocked != 2 {
		t.Fatalf("allowed=%d blocked=%d, want allowed=1 blocked=2", allowed, blocked)
	}
}

func TestSafeBrowserProxyBlocksDNSRebindingBeforeDial(t *testing.T) {
	resolver := &sequenceWebsiteResolver{answers: [][]net.IPAddr{
		{{IP: net.ParseIP("93.184.216.34")}},
		{{IP: net.ParseIP("127.0.0.1")}},
	}}
	dialer := &recordingWebsiteDialer{err: errors.New("dial should not run")}
	proxy := newSafeBrowserProxy(resolver, dialer)
	allowed := false
	if err := guardBrowserRequest(
		context.Background(),
		resolver,
		"https://restaurant.example/",
		func(context.Context) error { allowed = true; return nil },
		func(context.Context) error { return errors.New("unexpected block") },
	); err != nil {
		t.Fatalf("initial request guard: %v", err)
	}
	if !allowed {
		t.Fatal("initial public DNS answer was not allowed")
	}
	if _, err := proxy.dialPublicHost(context.Background(), "tcp", "restaurant.example", "443"); err == nil {
		t.Fatal("rebound private DNS answer was accepted")
	}
	if dialer.calls != 0 {
		t.Fatalf("dialer called %d times after private rebound, want 0", dialer.calls)
	}
}

func TestSafeBrowserProxyPinsValidatedIPAddress(t *testing.T) {
	resolver := staticWebsiteResolver{answers: map[string][]net.IPAddr{
		"restaurant.example": {{IP: net.ParseIP("93.184.216.34")}},
	}}
	dialer := &recordingWebsiteDialer{err: errors.New("expected test stop")}
	proxy := newSafeBrowserProxy(resolver, dialer)
	_, _ = proxy.dialPublicHost(context.Background(), "tcp", "restaurant.example", "443")
	if dialer.calls != 1 {
		t.Fatalf("dialer calls=%d, want 1", dialer.calls)
	}
	if dialer.address != "93.184.216.34:443" {
		t.Fatalf("dialer address=%q, want pinned public IP", dialer.address)
	}
}

func TestSafeBrowserProxyRejectsPrivateHTTPDestination(t *testing.T) {
	resolver := staticWebsiteResolver{answers: map[string][]net.IPAddr{
		"private.example": {{IP: net.ParseIP("10.0.0.5")}},
	}}
	dialer := &recordingWebsiteDialer{err: errors.New("dial should not run")}
	proxy := newSafeBrowserProxy(resolver, dialer)
	request := httptest.NewRequest(http.MethodGet, "http://private.example/internal", nil)
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want %d", response.Code, http.StatusForbidden)
	}
	if dialer.calls != 0 {
		t.Fatalf("dialer called %d times, want 0", dialer.calls)
	}
}
