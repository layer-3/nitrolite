package sign

import (
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// EthereumRawSigner is the implementation of the Signer interface for Ethereum Msg signing.
type EthereumMsgSigner struct {
	Signer
}

// Sign expects the input data to be a hash (e.g., Keccak256 hash).
func (s *EthereumMsgSigner) Sign(hash []byte) (Signature, error) {
	sig, err := s.Signer.Sign(ComputeEthereumSignedMessageHash(hash))
	if err != nil {
		return nil, err
	}
	return Signature(sig), nil
}

// NewEthereumMsgSigner creates a new Ethereum signer from a hex-encoded private key.
func NewEthereumMsgSigner(privateKeyHex string) (*EthereumMsgSigner, error) {
	signer, err := NewEthereumRawSigner(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return NewEthereumMsgSignerFromRaw(signer)
}

// NewEthereumMsgSignerFromRaw creates a new Ethereum signer from an existing Signer instance.
func NewEthereumMsgSignerFromRaw(signer Signer) (*EthereumMsgSigner, error) {
	return &EthereumMsgSigner{
		signer,
	}, nil
}

// EthereumAddressRecoverer implements the AddressRecoverer interface for Ethereum.
type EthereumMsgAddressRecoverer struct{}

// RecoverAddress implements the AddressRecoverer interface.
func (r *EthereumMsgAddressRecoverer) RecoverAddress(message []byte, signature Signature) (Address, error) {
	hash := ComputeEthereumSignedMessageHash(message)
	return RecoverAddressFromHash(hash, signature)
}

// ComputeEthereumSignedMessageHash accepts an arbitrary message, prepends a known message,
// and hashes the result using keccak256. The known message added to the input before hashing is
// "\x19Ethereum Signed Message:\n" + len(message).
func ComputeEthereumSignedMessageHash(message []byte) []byte {
	return ethcrypto.Keccak256(
		[]byte(
			fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), string(message)),
		),
	)
}
