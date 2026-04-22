// Package password centralises password handling for the auth service.
//
// All passwords handled by auth move through this package: it is the single
// place where the bcrypt cost is chosen and where hashes are verified. Keeping
// the wrapper narrow means the Dex-compatible format can be swapped in one
// spot if we ever move away from bcrypt.
package password

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Cost is the bcrypt work factor used for all hashes we produce. 10 is the
// Dex default and matches the example compose fixtures.
const Cost = 10

// ErrEmpty is returned when Hash is called with an empty password.
var ErrEmpty = errors.New("password is empty")

// Hash produces a bcrypt hash suitable for Dex's staticPasswords entries.
func Hash(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrEmpty
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plaintext), Cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(h), nil
}

// Verify reports whether plaintext matches hash. It returns (false, nil) on a
// genuine mismatch and a non-nil error on a malformed hash.
func Verify(hash, plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return false, nil
	default:
		return false, fmt.Errorf("bcrypt verify: %w", err)
	}
}

// IsBcrypt reports whether the given string looks like a bcrypt hash. It is
// purely a heuristic used to avoid rehashing values already in bcrypt form.
func IsBcrypt(s string) bool {
	if len(s) < 4 {
		return false
	}
	return s[0] == '$' && (s[:4] == "$2a$" || s[:4] == "$2b$" || s[:4] == "$2y$")
}
