package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

// Binding each candidate directly reserves the selected port through startup.
func listenDev(ctx context.Context, host string, port int, automatic bool, out io.Writer, listen func(context.Context, string, string) (net.Listener, error)) (net.Listener, error) {
	initial := port
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		listener, err := listen(ctx, "tcp4", net.JoinHostPort(host, fmt.Sprint(port)))
		if err == nil {
			if port != initial {
				fmt.Fprintf(out, "Starting dev on port %d.\n", port)
			}
			return listener, nil
		}
		if !errors.Is(err, syscall.EADDRINUSE) {
			return nil, fmt.Errorf("unable to start dev on port %d: %w", port, err)
		}
		if !automatic || port == 0 {
			return nil, fmt.Errorf("unable to start dev on port %d: another process is using it. Stop that process or choose another port with --port", port)
		}
		if port == 65535 {
			return nil, fmt.Errorf("unable to start dev: all ports from %d to 65535 are occupied. Free a port or choose another with --port", initial)
		}
		fmt.Fprintf(out, "Unable to start on port %d: another process is using it. Trying port %d.\n", port, port+1)
		port++
	}
}
