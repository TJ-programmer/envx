package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"
)

const tokenVersion byte = 0x80

type Fernet struct {
	signingKey    []byte
	encryptionKey []byte
}

func NewFernet(key []byte) (*Fernet, error) {
	if len(key) != 32 {
		return nil, errors.New("fernet key must be 32 bytes")
	}
	return &Fernet{signingKey: key[:16], encryptionKey: key[16:]}, nil
}

func (f *Fernet) Encrypt(data []byte) ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(f.encryptionKey)
	if err != nil {
		return nil, err
	}
	padded := pkcs7Pad(data, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(time.Now().Unix()))

	basic := make([]byte, 0, 1+8+aes.BlockSize+len(ciphertext))
	basic = append(basic, tokenVersion)
	basic = append(basic, ts...)
	basic = append(basic, iv...)
	basic = append(basic, ciphertext...)

	mac := hmac.New(sha256.New, f.signingKey)
	mac.Write(basic)
	return append(basic, mac.Sum(nil)...), nil
}

func (f *Fernet) Decrypt(token []byte) ([]byte, error) {
	const overhead = 1 + 8 + aes.BlockSize + 32
	if len(token) < overhead {
		return nil, errors.New("token too short")
	}
	if token[0] != tokenVersion {
		return nil, errors.New("unsupported token version")
	}

	basic := token[:len(token)-32]
	expected := token[len(token)-32:]
	mac := hmac.New(sha256.New, f.signingKey)
	mac.Write(basic)
	if !hmac.Equal(mac.Sum(nil), expected) {
		return nil, errors.New("invalid token")
	}

	iv := basic[9:25]
	ciphertext := basic[25:]

	block, err := aes.NewCipher(f.encryptionKey)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ciphertext)
	return pkcs7Unpad(plain, aes.BlockSize)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+pad)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid padding")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize || pad > len(data) {
		return nil, errors.New("invalid padding")
	}
	for _, b := range data[len(data)-pad:] {
		if int(b) != pad {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-pad], nil
}
