package cli

import (
	"bytes"
	"context"
	"errors"
	"net"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

func TestDevListenerPortFallback(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "0.0.0.0"} {
		t.Run(host, func(t *testing.T) {
			var out bytes.Buffer
			var addresses []string
			bound, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer bound.Close()
			listen := func(_ context.Context, network, address string) (net.Listener, error) {
				if network != "tcp4" {
					t.Fatal(network)
				}
				addresses = append(addresses, address)
				if len(addresses) < 3 {
					return nil, &net.OpError{Op: "listen", Err: syscall.EADDRINUSE}
				}
				return bound, nil
			}
			got, err := listenDev(context.Background(), host, 3000, true, &out, listen)
			if err != nil || got != bound {
				t.Fatal(got, err)
			}
			if !reflect.DeepEqual(addresses, []string{host + ":3000", host + ":3001", host + ":3002"}) {
				t.Fatal(addresses)
			}
			for _, line := range []string{
				"Unable to start on port 3000: another process is using it. Trying port 3001.",
				"Unable to start on port 3001: another process is using it. Trying port 3002.",
				"Starting dev on port 3002.",
			} {
				if !strings.Contains(out.String(), line) {
					t.Fatal(out.String())
				}
			}
		})
	}
}

func TestDevListenerStopsOnOtherErrorsAndFixedPorts(t *testing.T) {
	for _, tt := range []struct {
		name      string
		port      int
		automatic bool
		cause     error
		want      string
	}{
		{"fixed", 3000, false, syscall.EADDRINUSE, "choose another port with --port"},
		{"permission", 3000, true, syscall.EACCES, "permission denied"},
		{"exhausted", 65535, true, syscall.EADDRINUSE, "all ports from 65535 to 65535 are occupied"},
		{"ephemeral", 0, false, syscall.EADDRINUSE, "another process is using it"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			var out bytes.Buffer
			_, err := listenDev(context.Background(), "127.0.0.1", tt.port, tt.automatic, &out, func(context.Context, string, string) (net.Listener, error) {
				calls++
				return nil, &net.OpError{Op: "listen", Err: tt.cause}
			})
			if err == nil || !strings.Contains(err.Error(), tt.want) || calls != 1 || out.Len() != 0 {
				t.Fatal(err, calls, out.String())
			}
			if tt.cause != syscall.EADDRINUSE && !errors.Is(err, tt.cause) {
				t.Fatal("lost underlying error", err)
			}
		})
	}
}

func TestDevListenerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var out bytes.Buffer
	calls := 0
	_, err := listenDev(ctx, "127.0.0.1", 3000, true, &out, func(context.Context, string, string) (net.Listener, error) {
		calls++
		cancel()
		return nil, syscall.EADDRINUSE
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal(err, calls)
	}
}

func TestDevListenerBindsNextPort(t *testing.T) {
	occupied, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	port := occupied.Addr().(*net.TCPAddr).Port
	if port == 65535 {
		t.Skip("ephemeral listener selected last port")
	}
	var out bytes.Buffer
	chosen, err := listenDev(context.Background(), "127.0.0.1", port, true, &out, (&net.ListenConfig{}).Listen)
	if err != nil {
		t.Fatal(err)
	}
	defer chosen.Close()
	if chosen.Addr().(*net.TCPAddr).Port <= port || !strings.Contains(out.String(), "another process is using it") {
		t.Fatal(chosen.Addr(), out.String())
	}
}
