package cli

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strings"

	"flea/internal/devaccess"
)

func devLANAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("could not detect network addresses: %w", err)
	}
	var addresses []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		entries, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("could not inspect network interface %s: %w", iface.Name, err)
		}
		for _, entry := range entries {
			ip, _, err := net.ParseCIDR(entry.String())
			if err == nil && ip.To4() != nil && devaccess.IsLANIPv4(ip) {
				addresses = append(addresses, ip.String())
			}
		}
	}
	slices.Sort(addresses)
	addresses = slices.Compact(addresses)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no private or Tailscale IPv4 address found; connect to your trusted network, then retry --lan")
	}
	return addresses, nil
}

func devLANHostnames() ([]string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("could not detect this computer's hostname: %w", err)
	}
	hostname = strings.TrimSuffix(strings.ToLower(hostname), ".")
	short := strings.Split(hostname, ".")[0]
	names := []string{hostname, short, short + ".local"}
	slices.Sort(names)
	return slices.Compact(names), nil
}
