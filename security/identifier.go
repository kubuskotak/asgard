// Package security is func library that implement security standard.
package security

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	entropy = rand.Reader
	now     = time.Now()
)

// GenID returns random id with lexicographically sortable identifier.
func GenID() (string, error) {
	id, err := ulid.New(ulid.Timestamp(now), entropy)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
