package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	fernet, err := NewFernet(key)
	if err != nil {
		t.Fatal(err)
	}

	token, err := fernet.Encrypt([]byte("hello envx"))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := fernet.Decrypt(token)
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != "hello envx" {
		t.Fatalf("got %q", plain)
	}
}

func TestEncryptIsRandomized(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	fernet, _ := NewFernet(key)

	first, _ := fernet.Encrypt([]byte("same"))
	second, _ := fernet.Encrypt([]byte("same"))
	if string(first) == string(second) {
		t.Fatal("two encryptions of the same value must differ (random IV)")
	}
}

func TestTamperedTokenFails(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	fernet, _ := NewFernet(key)
	token, _ := fernet.Encrypt([]byte("hello"))

	tampered := append([]byte(nil), token...)
	tampered[len(tampered)-1] ^= 0x01
	if _, err := fernet.Decrypt(tampered); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}

func TestPythonParity(t *testing.T) {
	keyB64 := "5sLN-HKvCeXTE-Zd9kPSHqSCB6lEyr_gcdFSN171VCQ="
	tokenB64 := "gAAAAABqfu_lK0JbiqzSZyFRl_XFwrVwp8R-paR8rGr07sc2xSrCm-DUkuHQ2Qq0_uej35wunsLX1nIc5HclXyBy4_y2n-V6PeRx42_7E1GRHj5bE-0T8H4="

	key, err := base64.URLEncoding.DecodeString(keyB64)
	if err != nil {
		t.Fatal(err)
	}
	token, err := base64.URLEncoding.DecodeString(tokenB64)
	if err != nil {
		t.Fatal(err)
	}

	fernet, err := NewFernet(key)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := fernet.Decrypt(token)
	if err != nil {
		t.Fatalf("failed to decrypt Python-produced token: %v", err)
	}
	if got := string(plain); got != "hello envx parity" {
		t.Fatalf("parity mismatch: got %q, want %q", got, "hello envx parity")
	}
}

func TestEncryptedPrefixDetected(t *testing.T) {
	if !strings.HasPrefix(EncryptedPrefix+"abc", "enc:") {
		t.Fatal("prefix mismatch")
	}
}
