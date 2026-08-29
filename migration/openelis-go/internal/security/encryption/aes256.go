// Package encryption ports org.jasypt.util.text.AES256TextEncryptor, the
// TextEncryptor bean SecurityConfig publishes.
//
// It is reachable from one place in the ported surface so far: site_information
// rows flagged `encrypted`, whose value is stored as ciphertext, decrypted by
// the service on every read, and re-encrypted on every write. No shipped row
// carries the flag, so the whole path is unreachable on stock data and a port
// that ignored it agreed with Java on every byte — right up until someone
// created such a row through the admin screen.
//
// The parameters were read out of jasypt with a JDK rather than guessed:
//
//	algorithm    PBEWithHMACSHA512AndAES_256
//	key          PBKDF2-HMAC-SHA512, 1000 iterations, 256-bit
//	cipher       AES-256-CBC, PKCS#5 padding
//	salt         16 random bytes, RandomSaltGenerator, included in the output
//	iv           16 random bytes, RandomIvGenerator, included in the output
//	wire         Base64( salt || iv || ciphertext )
//
// The layout is worth stating plainly because reading the jasypt source
// suggests the opposite. StandardPBEByteEncryptor prepends the salt, then
// prepends the IV to THAT result — which reads as iv-then-salt. It is not:
// decrypting jasypt's own output proves salt comes first. Measured, not
// inferred, in both directions.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen    = 16
	ivLen      = 16
	keyLen     = 32 // AES-256
	iterations = 1000
)

// DefaultPassword is Spring's fallback for `encryption.general.password`, the
// literal in @Value("${encryption.general.password:dev}").
//
// It is NOT what this deployment uses: volume/properties/common.properties sets
// the key to `kspass`, and a value encrypted under one password is unreadable
// under the other. That file is the deployment's, not the application's, which
// is why the port takes the password from its environment rather than baking
// one in.
const DefaultPassword = "dev"

// TextEncryptor holds the password the key is derived from.
type TextEncryptor struct {
	Password string
}

// ErrMalformed is returned for input that cannot be a jasypt ciphertext.
var ErrMalformed = errors.New("encryption: malformed ciphertext")

// Encrypt returns Base64(salt || iv || AES-256-CBC(plaintext)).
//
// Salt and IV are fresh on every call, so encrypting one value twice produces
// two different strings. Nothing may compare ciphertexts for equality — see
// the write path, which compares decrypted values instead.
func (e *TextEncryptor) Encrypt(plaintext string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	iv := make([]byte, ivLen)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key(salt))
	if err != nil {
		return "", err
	}
	padded := pad([]byte(plaintext), block.BlockSize())
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)

	return base64.StdEncoding.EncodeToString(append(append(salt, iv...), out...)), nil
}

// Decrypt reverses Encrypt.
func (e *TextEncryptor) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrMalformed
	}
	if len(raw) <= saltLen+ivLen {
		return "", ErrMalformed
	}
	salt, iv, body := raw[:saltLen], raw[saltLen:saltLen+ivLen], raw[saltLen+ivLen:]

	block, err := aes.NewCipher(e.key(salt))
	if err != nil {
		return "", err
	}
	if len(body)%block.BlockSize() != 0 {
		return "", ErrMalformed
	}
	out := make([]byte, len(body))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, body)

	return unpad(out, block.BlockSize())
}

func (e *TextEncryptor) key(salt []byte) []byte {
	return pbkdf2.Key([]byte(e.Password), salt, iterations, keyLen, sha512.New)
}

// pad is PKCS#5/PKCS#7: n bytes of value n, so a plaintext that already fills a
// block gains a whole extra one.
func pad(b []byte, blockSize int) []byte {
	n := blockSize - len(b)%blockSize
	out := make([]byte, len(b), len(b)+n)
	copy(out, b)
	for i := 0; i < n; i++ {
		out = append(out, byte(n))
	}
	return out
}

func unpad(b []byte, blockSize int) (string, error) {
	if len(b) == 0 {
		return "", ErrMalformed
	}
	n := int(b[len(b)-1])
	if n == 0 || n > blockSize || n > len(b) {
		return "", ErrMalformed
	}
	for _, c := range b[len(b)-n:] {
		if int(c) != n {
			return "", ErrMalformed
		}
	}
	return string(b[:len(b)-n]), nil
}
