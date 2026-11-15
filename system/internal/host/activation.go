package host

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
)

type listeners struct {
	http, https net.Listener
}

func (l *listeners) close() {
	if l.http != nil {
		_ = l.http.Close()
	}
	if l.https != nil {
		_ = l.https.Close()
	}
}

// RunSocketActivated requires systemd's two wildcard TCP listeners. There is no
// bind fallback: a broken managed installation must fail rather than own ports
// that would disappear on restart. Standalone Run keeps ordinary direct binding.
func RunSocketActivated(ctx context.Context, configPath string, logger *log.Logger) error {
	l, err := inheritedListeners()
	if err != nil {
		return err
	}
	defer l.close()
	config, handler, err := Load(configPath, logger)
	if err != nil {
		return err
	}
	if config.Listen != ":80" || (config.TLS != nil && config.TLS.Listen != ":443") {
		return fmt.Errorf("socket activation requires listen :80 and HTTPS on :443")
	}
	return serveWithListeners(ctx, config, handler, logger, nil, l)
}

func inheritedListeners() (_ *listeners, err error) {
	return listenInherited(80, 443)
}

// Port arguments let subprocess tests exercise the real descriptor handoff on
// unprivileged ports. Production always requires the standard web ports.
func listenInherited(httpPort, httpsPort int) (_ *listeners, err error) {
	pid, pidErr := strconv.Atoi(os.Getenv("LISTEN_PID"))
	count, countErr := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	// Do not let children mistake these descriptors for their own activation.
	for _, name := range []string{"LISTEN_PID", "LISTEN_FDS", "LISTEN_FDNAMES"} {
		_ = os.Unsetenv(name)
	}
	if pidErr != nil || pid != os.Getpid() || countErr != nil || count != 2 {
		return nil, fmt.Errorf("socket activation requires two listening sockets from systemd for this process")
	}
	l := &listeners{}
	defer func() {
		if err != nil {
			l.close()
		}
	}()
	// Mark both originals before doing anything that could launch a child.
	for fd := 3; fd < 5; fd++ {
		syscall.CloseOnExec(fd)
	}
	for fd := 3; fd < 5; fd++ {
		file := os.NewFile(uintptr(fd), "systemd-listener")
		accepting, e := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_ACCEPTCONN)
		var listener net.Listener
		if e == nil && accepting == 1 {
			listener, e = net.FileListener(file)
		} else if e == nil {
			e = fmt.Errorf("descriptor is not listening")
		}
		_ = file.Close() // FileListener owns a separate, close-on-exec descriptor.
		if e != nil {
			// Close the other original too when conversion fails on the first.
			for next := fd + 1; next < 5; next++ {
				_ = syscall.Close(next)
			}
			return nil, fmt.Errorf("systemd listening socket: %w", e)
		}
		address, ok := listener.Addr().(*net.TCPAddr)
		if ok && address.IP.IsUnspecified() {
			switch {
			case address.Port == httpPort && l.http == nil:
				l.http = listener
				continue
			case address.Port == httpsPort && l.https == nil:
				l.https = listener
				continue
			}
		}
		_ = listener.Close()
		for next := fd + 1; next < 5; next++ {
			_ = syscall.Close(next)
		}
		return nil, fmt.Errorf("systemd sockets must be distinct wildcard TCP listeners on ports 80 and 443")
	}
	return l, nil
}
