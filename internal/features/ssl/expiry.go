package ssl

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

func checkExpiry(path string) (bool, string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return true, "failed to read file"
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return true, "failed to decode PEM"
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return true, "failed to parse certificate"
	}

	timeLeft := time.Until(cert.NotAfter)
	if timeLeft < 0 {
		return true, "expired"
	}

	// Renew if less than 30 days left
	if timeLeft < 30*24*time.Hour {
		return true, fmt.Sprintf("less than 30 days left (%s)", timeLeft.Round(time.Hour))
	}

	return false, timeLeft.Round(time.Hour).String()
}
