// Package devaccess carries the dev CLI's explicit LAN request authorization.
// Production never installs this middleware. Headers cannot grant LAN access.
package devaccess

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type lanKey struct{}

// LAN reports that the CLI checked this request's actual peer and exact Host.
func LAN(r *http.Request) bool {
	allowed, _ := r.Context().Value(lanKey{}).(bool)
	return allowed
}

// WrapLAN admits only private IPv4 peers and the listener's explicitly selected
// addresses and hostname aliases, at the selected port. Loopback remains usable. Forwarded headers are never consulted.
func WrapLAN(next http.Handler, addresses, names []string, port int) http.Handler {
	hosts := make(map[string]bool)
	allowHost := func(host string) {
		hosts[net.JoinHostPort(host, strconv.Itoa(port))] = true
		// Browsers omit the default HTTP port from Host and Origin.
		if port == 80 {
			hosts[host] = true
		}
	}
	allowHost("localhost")
	allowHost("127.0.0.1")
	for _, address := range addresses {
		ip := net.ParseIP(address)
		if ip != nil && ip.To4() != nil && IsLANIPv4(ip) {
			allowHost(ip.String())
		}
	}
	for _, name := range names {
		allowHost(strings.ToLower(name))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(peer)
		if err != nil || ip == nil || ip.To4() == nil || !IsLANIPv4(ip) || !hosts[strings.ToLower(r.Host)] {
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "Use a network address or hostname shown by flea dev.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), lanKey{}, true)))
	})
}

// IsLANIPv4 includes private IPv4, loopback and RFC 6598 shared addresses used by Tailscale.
func IsLANIPv4(ip net.IP) bool {
	v4 := ip.To4()
	return v4 != nil && (ip.IsLoopback() || ip.IsPrivate() || (v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127))
}
