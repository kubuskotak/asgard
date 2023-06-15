// Package security is func library that implement security standard.
package security

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/rs/zerolog/log"
)

const defaultRandomAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	max := big.NewInt(int64(len(defaultRandomAlphabet)))
	for i := range b {
		n := GenerateRandomInt(max)
		if n < 0 {
			return nil, fmt.Errorf("random int is failed: %d", n)
		}
		b[i] = defaultRandomAlphabet[n]
	}

	return b, nil
}

// GenerateRandomInt generates a random integer between min and max.
func GenerateRandomInt(max *big.Int) int64 {
	i, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Err(err).Msg("random bytes is failed")
		return 0
	}
	return i.Int64()
}

// GenerateRandomString generates a random string of length n.
func GenerateRandomString(n int) string {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		log.Err(err).Msg("random bytes is failed")
		return ""
	}
	return string(b)
}

// GenerateRandomEmail generates a random email.
func GenerateRandomEmail(n int) string {
	return fmt.Sprintf("%s@email.com", GenerateRandomString(n))
}
