// Package tlsutil provides TLS certificate management, including
// self-signed certificate generation using the Go crypto/x509 stdlib.
package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// EnsureCerts checks that TLS certificate and key files exist on disk.
// If they are missing and autoGenerate is true, a self-signed certificate
// is generated. If missing and autoGenerate is false, an error is returned.
func EnsureCerts(certFile, keyFile string, autoGenerate bool) error {
	certExists := fileExists(certFile)
	keyExists := fileExists(keyFile)

	if certExists && keyExists {
		return nil
	}

	if !autoGenerate {
		return fmt.Errorf("TLS certificate or key missing (cert=%s key=%s) and auto_generate is disabled", certFile, keyFile)
	}

	return generateSelfSigned(certFile, keyFile)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func generateSelfSigned(certFile, keyFile string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate EC key: %w", err)
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return fmt.Errorf("generate serial number: %w", err)
	}

	// Backdate 1 hour for RTCless Pis with clock skew at boot
	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: "Pi-Star Dashboard",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "pi-star.local", "pistar.local"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("create certificate: %w", err)
	}

	// Create parent directories with restricted permissions
	if err := os.MkdirAll(filepath.Dir(certFile), 0700); err != nil {
		return fmt.Errorf("create cert directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyFile), 0700); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	// Write certificate (world-readable)
	certOut, err := os.OpenFile(certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("write cert file: %w", err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("encode cert PEM: %w", err)
	}

	// Write private key (owner-only)
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal EC key: %w", err)
	}
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("write key file: %w", err)
	}
	defer keyOut.Close()
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		return fmt.Errorf("encode key PEM: %w", err)
	}

	return nil
}
