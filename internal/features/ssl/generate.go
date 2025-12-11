package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"compose-init/internal/config"
)

func Apply(cfg *config.ProjectConfig) error {
	// Top-level
	for _, cert := range cfg.GenerateCert {
		if err := generateOne(cert); err != nil {
			return fmt.Errorf("failed to generate cert for %s: %w", cert.Domain, err)
		}
	}

	// Service-level
	for _, svc := range cfg.Services {
		for _, cert := range svc.GenerateCert {
			if err := generateOne(cert); err != nil {
				return fmt.Errorf("failed to generate cert for %s: %w", cert.Domain, err)
			}
		}
	}
	return nil
}

func generateOne(c config.CertConfig) error {
	// Defaults
	if c.CertName == "" {
		c.CertName = "server.crt"
	}
	if c.KeyName == "" {
		c.KeyName = "server.key"
	}
	if c.OutputDir == "" {
		c.OutputDir = "./certs"
	}

	outDir, err := filepath.Abs(c.OutputDir)
	if err != nil {
		return err
	}

	certPath := filepath.Join(outDir, c.CertName)
	keyPath := filepath.Join(outDir, c.KeyName)

	if !c.Force {
		if _, err := os.Stat(certPath); err == nil {
			// Check expiry
			shouldRenew, reason := checkExpiry(certPath)
			if !shouldRenew {
				fmt.Printf("Certificate %s is valid (expires in %s), skipping.\n", certPath, reason)
				return nil
			}
			fmt.Printf("Certificate %s needs renewal: %s\n", certPath, reason)
		}
	} else {
		fmt.Printf("Forcing regeneration of %s\n", certPath)
	}

	fmt.Printf("Generating SSL cert for %s in %s\n", c.Domain, outDir)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Generate Key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Dev Cert"},
			CommonName:   c.Domain,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{c.Domain, "localhost"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Write Cert
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	// Write Key
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return nil
}
