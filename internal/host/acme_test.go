package host

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// This local ACME fixture implements the protocol needed for HTTP-01 issuance.
// It checks the served challenge and signs the actual CSR with a test-only CA.
// Production always uses Let's Encrypt; no insecure verification toggle exists.
func TestACMEIssuanceAndPersistentCache(t *testing.T) {
	c, h, _ := tlsFixture(t)
	c.TLS.CacheDir = t.TempDir()
	caKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	caTemplate := &x509.Certificate{SerialNumber: big.NewInt(10), Subject: pkix.Name{CommonName: "Local test CA"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(180 * 24 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, _ := x509.ParseCertificate(caDER)
	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	var mu sync.Mutex
	validated := false
	var chain []byte
	var nonce, issued atomic.Int32
	var base, challengeURL string
	ca := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Replay-Nonce", fmt.Sprintf("nonce-%d", nonce.Add(1)))
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		defer mu.Unlock()
		order := func() map[string]any {
			status := "pending"
			if validated {
				status = "ready"
			}
			if len(chain) > 0 {
				status = "valid"
			}
			return map[string]any{"status": status, "identifiers": []map[string]string{{"type": "dns", "value": "example.test"}}, "authorizations": []string{base + "/authz"}, "finalize": base + "/finalize", "certificate": base + "/cert"}
		}
		switch r.URL.Path {
		case "/directory":
			json.NewEncoder(w).Encode(map[string]any{"newNonce": base + "/nonce", "newAccount": base + "/account", "newOrder": base + "/new-order"})
		case "/nonce":
			w.WriteHeader(200)
		case "/account":
			w.Header().Set("Location", base+"/account/1")
			w.WriteHeader(201)
			io.WriteString(w, `{"status":"valid"}`)
		case "/new-order":
			w.Header().Set("Location", base+"/order")
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(order())
		case "/order":
			json.NewEncoder(w).Encode(order())
		case "/authz":
			status := "pending"
			if validated {
				status = "valid"
			}
			json.NewEncoder(w).Encode(map[string]any{"status": status, "identifier": map[string]string{"type": "dns", "value": "example.test"}, "challenges": []map[string]string{{"type": "http-01", "url": base + "/challenge", "token": "test-token", "status": status}}})
		case "/challenge":
			req, _ := http.NewRequest("GET", challengeURL+"/.well-known/acme-challenge/test-token", nil)
			req.Host = "example.test"
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				http.Error(w, "challenge failed", 500)
				return
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode != 200 || !strings.HasPrefix(string(body), "test-token.") {
				t.Errorf("challenge: %d %s", res.StatusCode, body)
				http.Error(w, "challenge failed", 500)
				return
			}
			validated = true
			json.NewEncoder(w).Encode(map[string]string{"status": "valid", "type": "http-01", "url": base + "/challenge", "token": "test-token"})
		case "/finalize":
			var jws struct{ Payload string }
			if err := json.NewDecoder(r.Body).Decode(&jws); err != nil {
				t.Error(err)
				http.Error(w, "bad jws", 400)
				return
			}
			payload, _ := base64.RawURLEncoding.DecodeString(jws.Payload)
			var data struct {
				CSR string `json:"csr"`
			}
			json.Unmarshal(payload, &data)
			csrDER, _ := base64.RawURLEncoding.DecodeString(data.CSR)
			csr, err := x509.ParseCertificateRequest(csrDER)
			if err != nil || csr.CheckSignature() != nil || !validated {
				t.Error("invalid CSR or authorization", err)
				http.Error(w, "bad csr", 400)
				return
			}
			leaf := &x509.Certificate{SerialNumber: big.NewInt(11), DNSNames: csr.DNSNames, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(90 * 24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
			der, err := x509.CreateCertificate(rand.Reader, leaf, caCert, csr.PublicKey, caKey)
			if err != nil {
				t.Error(err)
				http.Error(w, "sign failed", 500)
				return
			}
			chain = append(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})...)
			issued.Add(1)
			json.NewEncoder(w).Encode(order())
		case "/cert":
			w.Header().Set("Content-Type", "application/pem-certificate-chain")
			w.Write(chain)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ca.Close()
	base = ca.URL
	manager := &autocert.Manager{Prompt: autocert.AcceptTOS, Email: c.TLS.Email, Cache: autocert.DirCache(c.TLS.CacheDir), HostPolicy: autocert.HostWhitelist(c.TLS.Domains...), Client: &acme.Client{DirectoryURL: base + "/directory"}}
	httpHandler, tc, err := managedTransports(c, h, manager)
	if err != nil {
		t.Fatal(err)
	}
	challenge := httptest.NewServer(httpHandler)
	defer challenge.Close()
	challengeURL = challenge.URL
	request := func(tc *tls.Config) {
		t.Helper()
		s := httptest.NewUnstartedServer(tlsHandler(c, h))
		s.TLS = tc
		s.StartTLS()
		defer s.Close()
		tr := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, ServerName: "example.test"}}
		defer tr.CloseIdleConnections()
		client := &http.Client{Transport: tr, Timeout: 15 * time.Second}
		req, _ := http.NewRequest("GET", s.URL+"/", nil)
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
	request(tc)
	if issued.Load() != 1 {
		t.Fatal("no certificate issued", issued.Load())
	}
	info, err := os.Stat(filepath.Join(c.TLS.CacheDir, "example.test"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("certificate cache missing or permissions wrong", err)
	}
	ca.Close()
	// A fresh manager must be able to serve with the persisted certificate while
	// the CA is offline. This uses the production manager construction.
	_, cached, err := transports(c, h)
	if err != nil {
		t.Fatal(err)
	}
	request(cached)
}
