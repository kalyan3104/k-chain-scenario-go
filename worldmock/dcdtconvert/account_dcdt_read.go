package dcdtconvert

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
)

// MockDCDTData groups together all instances of a token (same token name, different nonces).
type MockDCDTData struct {
	TokenIdentifier []byte
	Instances       []*dcdt.DCDigitalToken
	LastNonce       uint64
	Roles           [][]byte
}

const (
	dcdtIdentifierSeparator  = "-"
	dcdtRandomSequenceLength = 6
)

// GetTokenBalance returns the DCDT balance of the account, specified by the
// token key.
func GetTokenBalance(tokenIdentifier []byte, nonce uint64, source map[string][]byte) (*big.Int, error) {
	tokenData, err := GetTokenData(tokenIdentifier, nonce, source, make(map[string][]byte))
	if err != nil {
		return nil, err
	}

	return tokenData.Value, nil
}

// GetTokenData gets the DCDT information related to a token from the storage of the account.
func GetTokenData(tokenIdentifier []byte, nonce uint64, source map[string][]byte, systemAccStorage map[string][]byte) (*dcdt.DCDigitalToken, error) {
	tokenKey := makeTokenKey(tokenIdentifier, nonce)
	return getTokenDataByKey(tokenKey, source, systemAccStorage)
}

func getTokenDataByKey(tokenKey []byte, source map[string][]byte, systemAccStorage map[string][]byte) (*dcdt.DCDigitalToken, error) {
	// default value copied from the protocol
	dcdtData := &dcdt.DCDigitalToken{
		Value: big.NewInt(0),
	}

	marshaledData := source[string(tokenKey)]
	if len(marshaledData) == 0 {
		return dcdtData, nil
	}

	err := dcdtDataMarshalizer.Unmarshal(dcdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	marshaledData = systemAccStorage[string(tokenKey)]
	if len(marshaledData) == 0 {
		return dcdtData, nil
	}
	dcdtDataFromSystemAcc := &dcdt.DCDigitalToken{}
	err = dcdtDataMarshalizer.Unmarshal(dcdtDataFromSystemAcc, marshaledData)
	if err != nil {
		return nil, err
	}

	dcdtData.TokenMetaData = dcdtDataFromSystemAcc.TokenMetaData

	return dcdtData, nil
}

// GetTokenRoles returns the roles of the account for the specified tokenName.
func GetTokenRoles(tokenName []byte, source map[string][]byte) ([][]byte, error) {
	tokenRolesKey := makeTokenRolesKey(tokenName)
	tokenRolesData := &dcdt.DCDTRoles{
		Roles: make([][]byte, 0),
	}

	marshaledData := source[string(tokenRolesKey)]
	if len(marshaledData) == 0 {
		return tokenRolesData.Roles, nil
	}

	err := dcdtDataMarshalizer.Unmarshal(tokenRolesData, marshaledData)
	if err != nil {
		return nil, err
	}

	return tokenRolesData.Roles, nil

}

// GetFullMockDCDTData returns the information about all the DCDT tokens held by the account.
func GetFullMockDCDTData(source map[string][]byte, systemAccStorage map[string][]byte) (map[string]*MockDCDTData, error) {
	resultMap := make(map[string]*MockDCDTData)
	for key := range source {
		storageKeyBytes := []byte(key)
		if isTokenKey(storageKeyBytes) {
			tokenName, tokenInstance, err := loadMockDCDTDataInstance(storageKeyBytes, source, systemAccStorage)
			if err != nil {
				return nil, err
			}
			if tokenInstance.Value.Sign() > 0 {
				resultObj := getOrCreateMockDCDTData(tokenName, resultMap)
				resultObj.Instances = append(resultObj.Instances, tokenInstance)
			}
		} else if isNonceKey(storageKeyBytes) {
			tokenName := key[len(dcdtNonceKeyPrefix):]
			resultObj := getOrCreateMockDCDTData(tokenName, resultMap)
			resultObj.LastNonce = big.NewInt(0).SetBytes(source[key]).Uint64()
		} else if isRoleKey(storageKeyBytes) {
			tokenName := key[len(dcdtRoleKeyPrefix):]
			roles, err := GetTokenRoles([]byte(tokenName), source)
			if err != nil {
				return nil, err
			}
			resultObj := getOrCreateMockDCDTData(tokenName, resultMap)
			resultObj.Roles = roles
		}
	}

	return resultMap, nil
}

func extractTokenIdentifierAndNonceDCDTWipe(args []byte) ([]byte, uint64) {
	argsSplit := bytes.Split(args, []byte(dcdtIdentifierSeparator))
	if len(argsSplit) < 2 {
		return args, 0
	}

	if len(argsSplit[1]) <= dcdtRandomSequenceLength {
		return args, 0
	}

	identifier := []byte(fmt.Sprintf("%s-%s", argsSplit[0], argsSplit[1][:dcdtRandomSequenceLength]))
	nonce := big.NewInt(0).SetBytes(argsSplit[1][dcdtRandomSequenceLength:])

	return identifier, nonce.Uint64()
}

// loads and prepared the DCDT instance
func loadMockDCDTDataInstance(tokenKey []byte, source map[string][]byte, systemAccStorage map[string][]byte) (string, *dcdt.DCDigitalToken, error) {
	tokenInstance, err := getTokenDataByKey(tokenKey, source, systemAccStorage)
	if err != nil {
		return "", nil, err
	}

	tokenNameFromKey := getTokenNameFromKey(tokenKey)
	tokenName, nonce := extractTokenIdentifierAndNonceDCDTWipe(tokenNameFromKey)

	if tokenInstance.TokenMetaData == nil {
		tokenInstance.TokenMetaData = &dcdt.MetaData{
			Name:  tokenName,
			Nonce: nonce,
		}
	}

	return string(tokenName), tokenInstance, nil
}

func getOrCreateMockDCDTData(tokenName string, resultMap map[string]*MockDCDTData) *MockDCDTData {
	resultObj := resultMap[tokenName]
	if resultObj == nil {
		resultObj = &MockDCDTData{
			TokenIdentifier: []byte(tokenName),
			Instances:       nil,
			LastNonce:       0,
			Roles:           nil,
		}
		resultMap[tokenName] = resultObj
	}
	return resultObj
}
