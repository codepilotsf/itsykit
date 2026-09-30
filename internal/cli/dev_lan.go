package cli

import (
	"fmt"
	"net"
	"slices"
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
			if err == nil && ip.To4() != nil && ip.IsPrivate() {
				addresses = append(addresses, ip.String())
			}
		}
	}
	slices.Sort(addresses)
	addresses = slices.Compact(addresses)
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no private IPv4 network address found; connect to Wi-Fi or Ethernet, then retry --lan")
	}
	return addresses, nil
}
