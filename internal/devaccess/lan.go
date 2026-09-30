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
// addresses. Loopback remains usable. Forwarded headers are never consulted.
func WrapLAN(next http.Handler, addresses []string, port int) http.Handler {
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
		if ip != nil && ip.To4() != nil && ip.IsPrivate() {
			allowHost(ip.String())
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(peer)
		if err != nil || ip == nil || ip.To4() == nil || (!ip.IsLoopback() && !ip.IsPrivate()) || !hosts[strings.ToLower(r.Host)] {
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "Use a network address shown by flea dev.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), lanKey{}, true)))
	})
}
