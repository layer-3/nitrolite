package sign_test

import (
	"fmt"
	"log"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/sign"
)

// ExampleNewEthereumRawSigner demonstrates creating an Ethereum signer and signing a message.
func ExampleNewEthereumRawSigner() {
	pkHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" // Example private key

	// Create a new Ethereum signer. It returns the generic sign.Signer interface.
	signer, err := sign.NewEthereumRawSigner(pkHex)
	if err != nil {
		log.Fatal(err)
	}

	// You can now use the signer for generic operations.
	fmt.Println("Address:", signer.PublicKey().Address())

	message := []byte("hello world")
	hash := ethcrypto.Keccak256Hash(message)
	signature, err := signer.Sign(hash.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Signature length:", len(signature))
	// Output:
	// Address: 0x1Be31A94361a391bBaFB2a4CCd704F57dc04d4bb
	// Signature length: 65
}

// ExampleSignature_String demonstrates the String method of Signature.
func ExampleSignature_String() {
	sig := sign.Signature([]byte{0x01, 0x02, 0x03, 0x04})
	fmt.Println(sig.String())
	// Output:
	// 0x01020304
}

// ExampleRecoverAddressFromHash demonstrates Ethereum-specific address recovery.
func ExampleRecoverAddressFromHash() {
	// Example message for standard recovery
	message := []byte("hello world")

	// Create a signature using our signer
	pkHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	signer, err := sign.NewEthereumRawSigner(pkHex)
	if err != nil {
		log.Fatal(err)
	}

	hash := ethcrypto.Keccak256Hash(message)
	signature, err := signer.Sign(hash.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	// Call the function directly from the `sign` package for hash recovery
	recoveredAddr, err := sign.RecoverAddressFromHash(hash.Bytes(), signature)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Verify it matches the signer's address
	signerAddr := signer.PublicKey().Address()
	fmt.Printf("Addresses match: %t\n", recoveredAddr.Equals(signerAddr))
	// Output:
	// Addresses match: true
}

// ExampleEthereumAddressRecoverer demonstrates using the generic AddressRecoverer interface.
func ExampleEthereumRawAddressRecoverer() {
	message := []byte("hello world")

	// Create a signer
	pkHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	signer, err := sign.NewEthereumRawSigner(pkHex)
	if err != nil {
		log.Fatal(err)
	}

	// Sign the message (note: Ethereum signers expect a hash)
	hash := ethcrypto.Keccak256Hash(message)
	signature, err := signer.Sign(hash.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	// Use the dedicated AddressRecoverer implementation
	var recoverer sign.AddressRecoverer = &sign.EthereumRawAddressRecoverer{}
	// The recoverer implementation will hash the raw message internally
	recoveredAddr, err := recoverer.RecoverAddress(message, signature)
	if err != nil {
		log.Fatal(err)
	}

	signerAddr := signer.PublicKey().Address()
	fmt.Printf("Generic recovery works: %t\n", recoveredAddr.Equals(signerAddr))
	// Output:
	// Generic recovery works: true
}
