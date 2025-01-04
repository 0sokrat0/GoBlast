// pkg/encryption/encryption_test.go

package encryption

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 байта
	plaintext := []byte("testtoken")

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Fatalf("Decrypted text does not match plaintext. Got: %s, Want: %s", decrypted, plaintext)
	}
}
