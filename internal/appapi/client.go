package appapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

var ErrUnavailable = fmt.Errorf("This app is temporarily unavailable. Please try again later.")
var totalCalls = make(chan struct{}, 64)

type Client struct {
	base                      url.URL
	name, token               string
	http                      *http.Client
	slots                     chan struct{}
	ready                     atomic.Bool
	timeout, handshakeTimeout time.Duration
}

// New accepts an administrator-provisioned service origin, never visitor data.
// Credentials are supplied separately, not included in URLs or site JSON.
func New(endpoint, name, token string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Hostname() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
		return nil, fmt.Errorf("app URL must be a service origin without credentials, path, query or fragment")
	}
	ip := net.ParseIP(u.Hostname())
	if u.Scheme != "https" && !(u.Scheme == "http" && ip != nil && ip.IsLoopback()) {
		return nil, fmt.Errorf("app URL requires HTTPS or an explicit loopback IP")
	}
	if !ValidName(name) || len(token) < 32 || len(token) > 256 || strings.IndexFunc(token, func(r rune) bool { return r < 33 || r > 126 }) >= 0 {
		return nil, fmt.Errorf("app name or private credential is invalid")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DisableCompression = true
	transport.MaxConnsPerHost = 4
	transport.MaxIdleConnsPerHost = 4
	transport.MaxResponseHeaderBytes = 16 << 10
	return &Client{base: *u, name: name, token: token, slots: make(chan struct{}, 4), timeout: 15 * time.Second, handshakeTimeout: 2 * time.Second, http: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func (c *Client) Close() { c.ready.Store(false); c.http.CloseIdleConnections() }

func (c *Client) call(ctx context.Context, method, route, query string, payload []byte, limit int64, timeout time.Duration, sessionToken string) (int, []byte, error) {
	select {
	case c.slots <- struct{}{}:
		defer func() { <-c.slots }()
	default:
		return 0, nil, ErrUnavailable
	}
	select {
	case totalCalls <- struct{}{}:
		defer func() { <-totalCalls }()
	default:
		return 0, nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	u := c.base
	u.Path = route
	u.RawQuery = query
	var body io.Reader
	// Use a non-rewindable reader: the transport must not retry a POST.
	if payload != nil {
		body = io.LimitReader(bytes.NewReader(payload), int64(len(payload)))
	}
	r, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return 0, nil, ErrUnavailable
	}
	r.Header.Set("Authorization", "Bearer "+c.token)
	r.Header.Set("Accept", "application/json")
	// The app's own opaque session token rides as trusted metadata, never as a
	// browser Cookie or Authorization header, and only when it is well formed.
	if ValidSessionToken(sessionToken) {
		r.Header.Set(SessionHeader, sessionToken)
	}
	if payload != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(r)
	if err != nil {
		return 0, nil, ErrUnavailable
	}
	defer response.Body.Close()
	kind, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || kind != "application/json" || response.Header.Get("Content-Encoding") != "" {
		return 0, nil, ErrUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(data)) > limit {
		return 0, nil, ErrUnavailable
	}
	return response.StatusCode, data, nil
}

func (c *Client) Handshake(ctx context.Context) error {
	_, err := c.HandshakeIdentity(ctx)
	return err
}

// HandshakeIdentity verifies and returns the running service's declared
// identity, so a caller can also compare it with the package manifest.
func (c *Client) HandshakeIdentity(ctx context.Context) (Identity, error) {
	c.ready.Store(false)
	status, body, err := c.call(ctx, "GET", "/_flea/handshake", "", nil, HandshakeLimit, c.handshakeTimeout, "")
	var identity Identity
	if err != nil || status != 200 || decode(body, &identity) != nil || identity.Name != c.name || identity.Protocol != Version || identity.Version == "" {
		return Identity{}, ErrUnavailable
	}
	c.ready.Store(true)
	return identity, nil
}

// Request returns only validated envelopes and mount-confined redirects.
func (c *Client) Request(ctx context.Context, mount Mount, method, publicPath, query string, fields map[string]any) (Envelope, int, error) {
	return c.RequestWithSession(ctx, mount, method, publicPath, query, fields, "")
}

// RequestWithSession is Request plus the app-owned session token read from
// Flea's own cookie. The token is validated again here before it is forwarded.
func (c *Client) RequestWithSession(ctx context.Context, mount Mount, method, publicPath, query string, fields map[string]any, sessionToken string) (Envelope, int, error) {
	var empty Envelope
	if !ValidSessionToken(sessionToken) {
		sessionToken = ""
	}
	if !c.ready.Load() || mount.Name != c.name || len(query) > FormLimit || (method != "GET" && method != "POST") {
		return empty, 503, ErrUnavailable
	}
	route, err := mount.ServicePath(publicPath)
	if err != nil {
		return empty, 503, ErrUnavailable
	}
	var payload []byte
	if method == "POST" {
		if fields == nil {
			fields = map[string]any{}
		}
		if _, exists := fields["_csrf"]; exists {
			return empty, 503, ErrUnavailable
		}
		for _, value := range fields {
			switch value.(type) {
			case string, []string:
			default:
				return empty, 503, ErrUnavailable
			}
		}
		payload, err = json.Marshal(fields)
		if err != nil || len(payload) > 6*FormLimit {
			return empty, 503, ErrUnavailable
		}
	} else if fields != nil {
		return empty, 503, ErrUnavailable
	}
	status, body, err := c.call(ctx, method, route, query, payload, ResponseLimit, c.timeout, sessionToken)
	var envelope Envelope
	if err != nil || decode(body, &envelope) != nil {
		return empty, 503, ErrUnavailable
	}
	// Explicit null has no protocol meaning; object/string fields must have the
	// documented shape. Reject unsupported instructions even when empty.
	var raw map[string]json.RawMessage
	if json.Unmarshal(body, &raw) != nil {
		return empty, 503, ErrUnavailable
	}
	for key, value := range raw {
		// session:null is meaningful; SessionDirective validates it below.
		if key == "session" {
			continue
		}
		if string(value) == "null" {
			return empty, 503, ErrUnavailable
		}
		if key == "flash" && len(value) > 4096 {
			return empty, 503, ErrUnavailable
		}
		if key == "redirect" {
			if _, err := mount.Redirect(envelope.Redirect); err != nil || status != 200 || method != "POST" {
				return empty, 503, ErrUnavailable
			}
		}
	}
	if status != 200 && status != 422 && status != 503 {
		return empty, 503, ErrUnavailable
	}
	// Validate the session instruction before any cookie can change. A session
	// directive is accepted only on an accepted 200 response.
	directive, err := envelope.SessionDirective()
	if err != nil || (directive.Present && status != 200) {
		return empty, 503, ErrUnavailable
	}
	if (status == 200 && (len(envelope.Invalid) != 0 || envelope.Error != "")) || (status == 422 && (len(envelope.Invalid) == 0 || envelope.Error != "")) || (status == 503 && (envelope.Error == "" || len(envelope.Invalid) != 0)) {
		return empty, 503, ErrUnavailable
	}
	return envelope, status, nil
}
