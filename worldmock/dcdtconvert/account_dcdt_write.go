package dcdtconvert

import (
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
	"github.com/kalyan3104/k-chain-vm-common-go/builtInFunctions"
)

// MakeDCDTUserMetadataBytes creates metadata byte slice
func MakeDCDTUserMetadataBytes(frozen bool) []byte {
	metadata := &builtInFunctions.DCDTUserMetadata{
		Frozen: frozen,
	}

	return metadata.ToBytes()
}

// WriteScenariosDCDTToStorage writes the Scenarios DCDT data to the provided storage map
func WriteScenariosDCDTToStorage(dcdtData []*scenmodel.DCDTData, destination map[string][]byte) error {
	for _, scenDCDTData := range dcdtData {
		tokenIdentifier := scenDCDTData.TokenIdentifier.Value
		isFrozen := scenDCDTData.Frozen.Value > 0
		for _, instance := range scenDCDTData.Instances {
			tokenNonce := instance.Nonce.Value
			tokenKey := makeTokenKey(tokenIdentifier, tokenNonce)
			tokenBalance := instance.Balance.Value
			var uris [][]byte
			for _, jsonUri := range instance.Uris.Values {
				uris = append(uris, jsonUri.Value)
			}
			tokenData := &dcdt.DCDigitalToken{
				Value:      tokenBalance,
				Type:       uint32(core.Fungible),
				Properties: MakeDCDTUserMetadataBytes(isFrozen),
				TokenMetaData: &dcdt.MetaData{
					Name:       []byte{},
					Nonce:      tokenNonce,
					Creator:    instance.Creator.Value,
					Royalties:  uint32(instance.Royalties.Value),
					Hash:       instance.Hash.Value,
					URIs:       uris,
					Attributes: instance.Attributes.Value,
				},
			}
			err := setTokenDataByKey(tokenKey, tokenData, destination)
			if err != nil {
				return err
			}
		}
		err := SetLastNonce(tokenIdentifier, scenDCDTData.LastNonce.Value, destination)
		if err != nil {
			return err
		}
		err = SetTokenRolesAsStrings(tokenIdentifier, scenDCDTData.Roles, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetTokenData sets the DCDT information related to a token into the storage of the account.
func setTokenDataByKey(tokenKey []byte, tokenData *dcdt.DCDigitalToken, destination map[string][]byte) error {
	marshaledData, err := dcdtDataMarshalizer.Marshal(tokenData)
	if err != nil {
		return err
	}
	destination[string(tokenKey)] = marshaledData
	return nil
}

// SetTokenData sets the token data
func SetTokenData(tokenIdentifier []byte, nonce uint64, tokenData *dcdt.DCDigitalToken, destination map[string][]byte) error {
	tokenKey := makeTokenKey(tokenIdentifier, nonce)
	return setTokenDataByKey(tokenKey, tokenData, destination)
}

// SetTokenRoles sets the specified roles to the account, corresponding to the given tokenIdentifier.
func SetTokenRoles(tokenIdentifier []byte, roles [][]byte, destination map[string][]byte) error {
	tokenRolesKey := makeTokenRolesKey(tokenIdentifier)
	tokenRolesData := &dcdt.DCDTRoles{
		Roles: roles,
	}

	marshaledData, err := dcdtDataMarshalizer.Marshal(tokenRolesData)
	if err != nil {
		return err
	}

	destination[string(tokenRolesKey)] = marshaledData
	return nil
}

// SetTokenRolesAsStrings sets the specified roles to the account, corresponding to the given tokenIdentifier.
func SetTokenRolesAsStrings(tokenIdentifier []byte, rolesAsStrings []string, destination map[string][]byte) error {
	roles := make([][]byte, len(rolesAsStrings))
	for i := 0; i < len(roles); i++ {
		roles[i] = []byte(rolesAsStrings[i])
	}

	return SetTokenRoles(tokenIdentifier, roles, destination)
}

// SetLastNonce writes the last nonce of a specified DCDT into the storage.
func SetLastNonce(tokenIdentifier []byte, lastNonce uint64, destination map[string][]byte) error {
	tokenNonceKey := makeLastNonceKey(tokenIdentifier)
	nonceBytes := big.NewInt(0).SetUint64(lastNonce).Bytes()
	destination[string(tokenNonceKey)] = nonceBytes
	return nil
}

// SetTokenBalance sets the DCDT balance of the account, specified by the token
// key.
func SetTokenBalance(tokenIdentifier []byte, nonce uint64, balance *big.Int, destination map[string][]byte) error {
	tokenKey := makeTokenKey(tokenIdentifier, nonce)
	tokenData, err := getTokenDataByKey(tokenKey, destination, make(map[string][]byte))
	if err != nil {
		return err
	}

	if balance.Sign() < 0 {
		return errNegativeValue
	}

	tokenData.Value = balance
	return setTokenDataByKey(tokenKey, tokenData, destination)
}
