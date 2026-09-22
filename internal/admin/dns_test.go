package admin

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
)

func TestDNSPreflight(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ips     []string
		err     error
		wantErr bool
	}{
		{"matching dual stack", []string{"192.0.2.10", "2001:db8::10"}, nil, false},
		{"wrong A", []string{"192.0.2.11"}, nil, true},
		{"wrong AAAA", []string{"192.0.2.10", "2001:db8::11"}, nil, true},
		{"no records", nil, nil, true}, {"NXDOMAIN", nil, errors.New("no such host"), true}, {"timeout", nil, context.DeadlineExceeded, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkDNS(context.Background(), []string{"example.test"}, []string{"192.0.2.10", "2001:db8::10"}, io.Discard, func(context.Context, string) ([]net.IPAddr, error) {
				var ips []net.IPAddr
				for _, v := range tc.ips {
					ips = append(ips, net.IPAddr{IP: net.ParseIP(v)})
				}
				return ips, tc.err
			})
			if (err != nil) != tc.wantErr {
				t.Fatal(err)
			}
		})
	}
	if err := checkDNS(context.Background(), []string{"192.0.2.10"}, nil, io.Discard, func(context.Context, string) ([]net.IPAddr, error) { t.Fatal("looked up IP literal"); return nil, nil }); err != nil {
		t.Fatal(err)
	}
	seen := []string{}
	err := checkDNS(context.Background(), []string{"example.test", "www.example.test"}, []string{"192.0.2.10"}, io.Discard, func(_ context.Context, h string) ([]net.IPAddr, error) {
		seen = append(seen, h)
		return []net.IPAddr{{IP: net.ParseIP("192.0.2.10")}}, nil
	})
	if err != nil || len(seen) != 2 {
		t.Fatal(err, seen)
	}
}
