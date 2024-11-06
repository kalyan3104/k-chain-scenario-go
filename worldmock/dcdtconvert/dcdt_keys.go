package dcdtconvert

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/marshal"
)

// dcdtTokenKeyPrefix is the prefix of storage keys belonging to DCDT tokens.
var dcdtTokenKeyPrefix = []byte(core.ProtectedKeyPrefix + core.DCDTKeyIdentifier)

// dcdtRoleKeyPrefix is the prefix of storage keys belonging to DCDT roles.
var dcdtRoleKeyPrefix = []byte(core.ProtectedKeyPrefix + core.DCDTRoleIdentifier + core.DCDTKeyIdentifier)

// dcdtNonceKeyPrefix is the prefix of storage keys belonging to DCDT nonces.
var dcdtNonceKeyPrefix = []byte(core.ProtectedKeyPrefix + core.DCDTNFTLatestNonceIdentifier)

// dcdtDataMarshalizer is the global marshalizer to be used for encoding/decoding DCDT data
var dcdtDataMarshalizer = &marshal.GogoProtoMarshalizer{}

// errNegativeValue signals that a negative value has been detected and it is not allowed
var errNegativeValue = errors.New("negative value")

// makeTokenKey creates the storage key corresponding to the given tokenName.
func makeTokenKey(tokenName []byte, nonce uint64) []byte {
	nonceBytes := big.NewInt(0).SetUint64(nonce).Bytes()
	tokenKey := append(dcdtTokenKeyPrefix, tokenName...)
	tokenKey = append(tokenKey, nonceBytes...)
	return tokenKey
}

// makeTokenRolesKey creates the storage key corresponding to the roles for the
// given tokenName.
func makeTokenRolesKey(tokenName []byte) []byte {
	tokenRolesKey := append(dcdtRoleKeyPrefix, tokenName...)
	return tokenRolesKey
}

// makeLastNonceKey creates the storage key corresponding to the last nonce of
// the given tokenName.
func makeLastNonceKey(tokenName []byte) []byte {
	tokenNonceKey := append(dcdtNonceKeyPrefix, tokenName...)
	return tokenNonceKey
}

// isTokenKey returns true if the given storage key belongs to an DCDT token.
func isTokenKey(key []byte) bool {
	return bytes.HasPrefix(key, dcdtTokenKeyPrefix)
}

// isRoleKey returns true if the given storage key belongs to an DCDT role.
func isRoleKey(key []byte) bool {
	return bytes.HasPrefix(key, dcdtRoleKeyPrefix)
}

// isNonceKey returns true if the given storage key belongs to an DCDT nonce.
func isNonceKey(key []byte) bool {
	return bytes.HasPrefix(key, dcdtNonceKeyPrefix)
}

// getTokenNameFromKey extracts the token name from the given storage key; it
// does not check whether the key is indeed a token key or not.
func getTokenNameFromKey(key []byte) []byte {
	return key[len(dcdtTokenKeyPrefix):]
}
