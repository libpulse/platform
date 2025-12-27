package crypto

import (
	"testing"
)

// Initialize crypto for all tests in this package
func init() {
	Init("test-pepper-for-hmac-hashing")
}

func TestHashSecret(t *testing.T) {
	secret := "psk_live_test123"
	hash1 := HashSecret(secret)
	hash2 := HashSecret(secret)

	// Same input should produce same hash
	if hash1 != hash2 {
		t.Errorf("HashSecret not deterministic: %s != %s", hash1, hash2)
	}

	// Hash should be 64 characters (SHA256 hex output)
	if len(hash1) != 64 {
		t.Errorf("HashSecret output length = %d, want 64", len(hash1))
	}

	// Different inputs should produce different hashes
	secret2 := "psk_live_different"
	hash3 := HashSecret(secret2)
	if hash1 == hash3 {
		t.Errorf("HashSecret produced same hash for different inputs")
	}
}

func TestGetLast4(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"psk_live_abcdefg", "defg"},
		{"abc", "abc"},
		{"", ""},
		{"1234", "1234"},
	}

	for _, tt := range tests {
		result := GetLast4(tt.input)
		if result != tt.expected {
			t.Errorf("GetLast4(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGeneratePublicKey(t *testing.T) {
	key, err := GeneratePublicKey()
	if err != nil {
		t.Fatalf("GeneratePublicKey error: %v", err)
	}

	// Should start with pk_live_
	prefix := "pk_live_"
	if len(key) < len(prefix) || key[:len(prefix)] != prefix {
		t.Errorf("GeneratePublicKey = %q, should start with %q", key, prefix)
	}

	// Should be unique
	key2, _ := GeneratePublicKey()
	if key == key2 {
		t.Errorf("GeneratePublicKey produced duplicate keys")
	}
}

func TestGenerateSecret(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret error: %v", err)
	}

	// Should start with psk_live_
	prefix := "psk_live_"
	if len(secret) < len(prefix) || secret[:len(prefix)] != prefix {
		t.Errorf("GenerateSecret = %q, should start with %q", secret, prefix)
	}

	// Should be unique
	secret2, _ := GenerateSecret()
	if secret == secret2 {
		t.Errorf("GenerateSecret produced duplicate secrets")
	}
}
