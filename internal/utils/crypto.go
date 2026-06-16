// Package utils provides cryptographic helpers using stdlib only.
// PBKDF2-SHA256 at 310,000 iterations as per OWASP 2023 recommendations.
package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const (
	pbkdf2Iterations = 310_000
	pbkdf2SaltLen    = 32
	pbkdf2KeyLen     = 32
)

// HashPassword returns a hex-encoded string: salt$hash (both hex).
func HashPassword(password string) (string, error) {
	salt := make([]byte, pbkdf2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("crypto: rand: %w", err)
	}
	key := deriveKey([]byte(password), salt)
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(key), nil
}

// CheckPassword returns true if password matches the stored hash.
func CheckPassword(password, stored string) bool {
	if len(stored) < pbkdf2SaltLen*2+1 {
		return false
	}
	saltHex := stored[:pbkdf2SaltLen*2]
	hashHex := stored[pbkdf2SaltLen*2+1:]
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(hashHex)
	if err != nil {
		return false
	}
	got := deriveKey([]byte(password), salt)
	return hmac.Equal(got, expected)
}

// GenerateToken returns a cryptographically random 64-char hex string (32 bytes).
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// deriveKey runs PBKDF2-SHA256 using only stdlib hmac + sha256.
func deriveKey(password, salt []byte) []byte {
	prf := hmac.New(sha256.New, password)
	hashLen := prf.Size()
	numBlocks := (pbkdf2KeyLen + hashLen - 1) / hashLen
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	counter := make([]byte, 4)
	for block := 1; block <= numBlocks; block++ {
		binary.BigEndian.PutUint32(counter, uint32(block))
		prf.Reset()
		prf.Write(salt)
		prf.Write(counter)
		U = prf.Sum(U[:0])
		T := append([]byte{}, U...)
		for i := 1; i < pbkdf2Iterations; i++ {
			prf.Reset()
			prf.Write(U)
			U = prf.Sum(U[:0])
			for j := range T {
				T[j] ^= U[j]
			}
		}
		dk = append(dk, T...)
	}
	return dk[:pbkdf2KeyLen]
}
