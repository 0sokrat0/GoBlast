// pkg/encryption/encryption.go

package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Encrypt шифрует plaintext с использованием AES-GCM.
// Возвращает зашифрованные данные или ошибку.
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Создаём новый блок AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаём GCM (Galois/Counter Mode)
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Генерируем nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Шифруем данные
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt дешифрует ciphertext с использованием AES-GCM.
// Возвращает расшифрованные данные или ошибку.
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Создаём новый блок AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаём GCM (Galois/Counter Mode)
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Определяем размер nonce
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Извлекаем nonce и зашифрованные данные
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Дешифруем данные
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
