package sign

import (
	"bytes"
	"testing"
)

func TestMockSigner(t *testing.T) {
	signer := NewMockSigner("test-id")
	data := []byte("test data")

	// Test Sign
	sig, err := signer.Sign(data)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	expectedSig := []byte("test data-signed-by-test-id")
	if !bytes.Equal(sig, expectedSig) {
		t.Errorf("got signature %q, want %q", sig, expectedSig)
	}

	// Test PublicKey
	pk := signer.PublicKey()
	if pk.Address().String() != "test-id" {
		t.Errorf("got address %q, want %q", pk.Address().String(), "test-id")
	}
}

func TestMockPublicKey(t *testing.T) {
	pk := NewMockPublicKey("key-id")

	// Test Address
	if pk.Address().String() != "key-id" {
		t.Errorf("got address %q, want %q", pk.Address().String(), "key-id")
	}

	// Test Bytes
	if !bytes.Equal(pk.Bytes(), []byte("key-id")) {
		t.Errorf("got bytes %q, want %q", pk.Bytes(), []byte("key-id"))
	}
}

func TestMockAddress(t *testing.T) {
	addr1 := NewMockAddress("addr1")
	addr2 := NewMockAddress("addr1")
	addr3 := NewMockAddress("addr2")

	// Test String
	if addr1.String() != "addr1" {
		t.Errorf("got string %q, want %q", addr1.String(), "addr1")
	}

	// Test Equals
	if !addr1.Equals(addr2) {
		t.Error("addr1 should equal addr2")
	}

	if addr1.Equals(addr3) {
		t.Error("addr1 should not equal addr3")
	}
}
