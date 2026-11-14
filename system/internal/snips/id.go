// Package snips contains building blocks for Blog content.
package snips

import (
	"crypto/rand"
	"fmt"
	"io"
)

const IDLength = 10
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// ReserveID generates and reserves an ID in the supplied set of Blog posts.
// Callers persisting new content must hold
// their content-write lock through allocation and persistence; this function
// does not write files or synchronize access to the map.
func ReserveID(used map[string]struct{}) (string, error) {
	return reserveID(rand.Reader, used)
}

func reserveID(random io.Reader, used map[string]struct{}) (string, error) {
	if used == nil {
		return "", fmt.Errorf("Cannot create a content ID without the existing ID index.")
	}
	// Bound retries as well as entropy reads. Never fall back to predictable IDs.
	for attempt := 0; attempt < 128; attempt++ {
		var id [IDLength]byte
		length := 0
		for read := 0; length < IDLength && read < 1024; read++ {
			var b [1]byte
			if _, err := io.ReadFull(random, b[:]); err != nil {
				return "", fmt.Errorf("Could not generate a secure content ID: %w", err)
			}
			// 248 is the largest multiple of 62 below 256. Rejection avoids
			// modulo bias without adding a NanoID runtime or package dependency.
			if b[0] >= 248 {
				continue
			}
			id[length] = alphabet[int(b[0])%len(alphabet)]
			length++
		}
		if length != IDLength {
			return "", fmt.Errorf("Could not generate a secure content ID. Please try again.")
		}
		value := string(id[:])
		if _, exists := used[value]; exists {
			continue
		}
		used[value] = struct{}{}
		return value, nil
	}
	return "", fmt.Errorf("Could not allocate a unique content ID. No content was changed.")
}
