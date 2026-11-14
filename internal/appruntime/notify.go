package appruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	contactapp "itsy/apps/contact"
)

func notify(ctx context.Context, cfg launchConfig, s contactapp.Submission) error {
	u, err := url.Parse(cfg.CapabilityURL)
	if err != nil || u.Scheme != "http" || !net.ParseIP(u.Hostname()).IsLoopback() || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("mail capability is unavailable")
	}
	body, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("invalid notification")
	}
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(ctx, "POST", cfg.CapabilityURL+CapabilityPath, io.LimitReader(bytes.NewReader(body), int64(len(body))))
	if err != nil {
		return fmt.Errorf("mail capability is unavailable")
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+cfg.CapabilityToken)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	defer transport.CloseIdleConnections()
	c := &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := c.Do(r)
	if err != nil {
		return fmt.Errorf("could not confirm delivery")
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != 204 {
		return fmt.Errorf("could not confirm delivery")
	}
	return nil
}
