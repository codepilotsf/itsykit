package admin

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"itsy/internal/host"
)

func CheckDNS(ctx context.Context, domains, expected []string, out io.Writer) error {
	if len(expected) == 0 {
		addresses, err := net.InterfaceAddrs()
		if err != nil {
			return fmt.Errorf("read server interface addresses: %w", err)
		}
		for _, address := range addresses {
			ip, _, err := net.ParseCIDR(address.String())
			if err == nil && !ip.IsLoopback() && !ip.IsUnspecified() && !ip.IsLinkLocalUnicast() {
				expected = append(expected, ip.String())
			}
		}
	}
	resolver := &net.Resolver{PreferGo: true, StrictErrors: true}
	return checkDNS(ctx, domains, expected, out, resolver.LookupIPAddr)
}

func checkDNS(ctx context.Context, domains, expected []string, out io.Writer, lookup func(context.Context, string) ([]net.IPAddr, error)) error {
	allowed := map[string]bool{}
	for _, value := range expected {
		ip := net.ParseIP(value)
		if ip == nil {
			return fmt.Errorf("invalid expected address %q", value)
		}
		allowed[ip.String()] = true
	}
	for _, domain := range domains {
		if !host.ValidHost(domain) {
			return fmt.Errorf("invalid DNS hostname %q", domain)
		}
		if net.ParseIP(domain) != nil {
			fmt.Fprintf(out, "DNS: %s is an IP destination; no lookup required.\n", domain)
			continue
		}
		deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
		addresses, err := lookup(deadline, domain)
		cancel()
		if err != nil {
			return fmt.Errorf("DNS lookup for %s from this server failed: %w", domain, err)
		}
		if len(addresses) == 0 {
			return fmt.Errorf("DNS lookup for %s returned no A or AAAA addresses", domain)
		}
		for _, address := range addresses {
			if !allowed[address.IP.String()] {
				return fmt.Errorf("DNS for %s resolves to %s, not an expected server address (%s); fix DNS or configure expected_ips for NAT/proxy", domain, address.IP, strings.Join(expected, ", "))
			}
		}
		fmt.Fprintf(out, "DNS: %s resolves to expected addresses.\n", domain)
	}
	return nil
}
