package worldmock

import (
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
	"github.com/kalyan3104/k-chain-scenario-go/worldmock/dcdtconvert"
)

// GetTokenBalance returns the DCDT balance of the account, specified by the
// token key.
func (a *Account) GetTokenBalance(tokenIdentifier []byte, nonce uint64) (*big.Int, error) {
	return dcdtconvert.GetTokenBalance(tokenIdentifier, nonce, a.Storage)
}

// GetTokenBalanceUint64 returns the DCDT balance of the account, specified by the
// token key.
func (a *Account) GetTokenBalanceUint64(tokenIdentifier []byte, nonce uint64) (uint64, error) {
	balance, err := a.GetTokenBalance(tokenIdentifier, nonce)
	if err != nil {
		return 0, err
	}
	return balance.Uint64(), nil
}

// SetTokenBalance sets the DCDT balance of the account, specified by the token
// key.
func (a *Account) SetTokenBalance(tokenIdentifier []byte, nonce uint64, balance *big.Int) error {
	return dcdtconvert.SetTokenBalance(tokenIdentifier, nonce, balance, a.Storage)
}

// SetTokenBalanceUint64 sets the DCDT balance of the account, specified by the
// token key.
func (a *Account) SetTokenBalanceUint64(tokenIdentifier []byte, nonce uint64, balance uint64) error {
	return dcdtconvert.SetTokenBalance(tokenIdentifier, nonce, big.NewInt(0).SetUint64(balance), a.Storage)
}

// GetTokenData gets the DCDT information related to a token from the storage of the account.
func (a *Account) GetTokenData(tokenIdentifier []byte, nonce uint64, systemAccStorage map[string][]byte) (*dcdt.DCDigitalToken, error) {
	return dcdtconvert.GetTokenData(tokenIdentifier, nonce, a.Storage, systemAccStorage)
}

// SetTokenData sets the DCDT information related to a token into the storage of the account.
func (a *Account) SetTokenData(tokenIdentifier []byte, nonce uint64, tokenData *dcdt.DCDigitalToken) error {
	return dcdtconvert.SetTokenData(tokenIdentifier, nonce, tokenData, a.Storage)
}

// SetTokenRolesAsStrings sets the specified roles to the account, corresponding to the given tokenName.
func (a *Account) SetTokenRolesAsStrings(tokenIdentifier []byte, rolesAsStrings []string) error {
	return dcdtconvert.SetTokenRolesAsStrings(tokenIdentifier, rolesAsStrings, a.Storage)
}
