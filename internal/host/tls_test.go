package host

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func tlsFixture(t *testing.T) (Config, http.Handler, *x509.CertPool) {
	t.Helper()
	c, h, err := Load(testConfig(t, ":80"), nil)
	if err != nil {
		t.Fatal(err)
	}
	c.TLS = &TLSConfig{Listen: ":443", Email: "dev@example.test", AgreeTOS: true, CacheDir: t.TempDir(), Domains: []string{"example.test"}}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "example.test"}, DNSNames: []string{"example.test"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(90 * 24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, BasicConstraintsValid: true, IsCA: true}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	private, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	data := append(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: private}), certPEM...)
	if err := os.WriteFile(filepath.Join(c.TLS.CacheDir, "example.test"), data, 0600); err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(certPEM)
	return c, h, pool
}

func TestHTTPSRedirectAndValidation(t *testing.T) {
	c, h, _ := tlsFixture(t)
	if err := validateTLS(c); err != nil {
		t.Fatal(err)
	}
	redirect := redirectHTTPS(c, h)
	for _, tc := range []struct {
		host, path string
		status     int
		location   string
	}{
		{"EXAMPLE.TEST:80", "/a%2Fb?q=1", 308, "https://example.test/a%2Fb?q=1"},
		{"evil.test", "/", 404, ""}, {"example.test", HealthPath, 200, ""},
	} {
		r := httptest.NewRequest("GET", tc.path, nil)
		r.Host = tc.host
		w := httptest.NewRecorder()
		redirect.ServeHTTP(w, r)
		if w.Code != tc.status || w.Header().Get("Location") != tc.location {
			t.Fatal(w.Code, w.Header())
		}
	}
	for _, domain := range []string{"127.0.0.1", "*.example.com", "bad/path", "localhost", "UPPER.example", "bad..example", "-bad.example", "example.com."} {
		if ValidDomain(domain) {
			t.Fatal("accepted", domain)
		}
	}
	c.TLS.Domains = append(c.TLS.Domains, "unknown.test")
	if err := validateTLS(c); err == nil {
		t.Fatal("accepted unconfigured domain")
	}
	c.TLS.Domains = c.TLS.Domains[:1]
	c.TLS.AgreeTOS = false
	if err := validateTLS(c); err == nil {
		t.Fatal("accepted missing consent")
	}
}

func TestCachedCertificateHTTPSAndHostIsolation(t *testing.T) {
	c, h, pool := tlsFixture(t)
	_, tc, err := transports(c, h)
	if err != nil {
		t.Fatal(err)
	}
	// Exercise autocert's real cache loader with a certificate trusted only by
	// this test client. No network access to a CA is needed.
	s := httptest.NewUnstartedServer(tlsHandler(c, h))
	s.TLS = tc
	s.StartTLS()
	defer s.Close()
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, ServerName: "example.test", MinVersion: tls.VersionTLS12}}
	client := &http.Client{Transport: tr}
	defer tr.CloseIdleConnections()
	for _, host := range []string{"example.test", "other.test"} {
		req, _ := http.NewRequest("GET", s.URL+"/", nil)
		req.Host = host
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		if host == "example.test" {
			if res.StatusCode != 200 || !strings.Contains(string(body), "Host test") {
				t.Fatal(res.StatusCode, string(body))
			}
		} else if res.StatusCode != 421 {
			t.Fatal("domain fronting allowed", res.StatusCode)
		}
	}
	if _, err := tc.GetCertificate(&tls.ClientHelloInfo{ServerName: "other.test"}); err == nil {
		t.Fatal("unknown SNI accepted")
	}
	if _, err := tc.GetCertificate(&tls.ClientHelloInfo{}); err == nil {
		t.Fatal("empty SNI accepted")
	}
}

func TestDualListenersAndShutdown(t *testing.T) {
	c, h, pool := tlsFixture(t)
	available := func() string {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		addr := l.Addr().String()
		l.Close()
		return addr
	}
	c.Listen = available()
	c.TLS.Listen = available()
	ctx, stop := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, c, h, nil, nil) }()
	defer func() {
		stop()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(6 * time.Second):
			t.Error("shutdown hung")
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", c.Listen, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, ServerName: "example.test"}}
	defer tr.CloseIdleConnections()
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", "https://"+c.TLS.Listen+"/", nil)
	req.Host = "example.test"
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal(res.StatusCode)
	}
}
