package scenexec

import (
	"fmt"
	"sort"
	"strings"

	oj "github.com/kalyan3104/k-chain-scenario-go/orderedjson"
	er "github.com/kalyan3104/k-chain-scenario-go/scenario/expression/reconstructor"
	scenjwrite "github.com/kalyan3104/k-chain-scenario-go/scenario/json/write"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
	worldmock "github.com/kalyan3104/k-chain-scenario-go/worldmock"
	"github.com/kalyan3104/k-chain-scenario-go/worldmock/dcdtconvert"

	"github.com/kalyan3104/k-chain-core-go/core"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
)

const includeProtectedStorage = false

func (ae *ScenarioExecutor) convertMockAccountToScenarioFormat(account *worldmock.Account) (*scenmodel.Account, error) {
	var storageKeys []string
	for storageKey := range account.Storage {
		storageKeys = append(storageKeys, storageKey)
	}

	sort.Strings(storageKeys)
	var storageKvps []*scenmodel.StorageKeyValuePair
	for _, storageKey := range storageKeys {
		storageValue := account.Storage[storageKey]
		includeKey := includeProtectedStorage || !strings.HasPrefix(storageKey, core.ProtectedKeyPrefix)
		if includeKey && len(storageValue) > 0 {
			storageKvps = append(storageKvps, &scenmodel.StorageKeyValuePair{
				Key: scenmodel.JSONBytesFromString{
					Value:    []byte(storageKey),
					Original: ae.exprReconstructor.Reconstruct([]byte(storageKey), er.NoHint),
				},
				Value: scenmodel.JSONBytesFromTree{
					Value:    storageValue,
					Original: &oj.OJsonString{Value: ae.exprReconstructor.Reconstruct(storageValue, er.NoHint)},
				},
			})
		}
	}

	systemAccStorage := make(map[string][]byte)
	systemAcc, exists := ae.World.AcctMap[string(vmcommon.SystemAccountAddress)]
	if exists {
		systemAccStorage = systemAcc.Storage
	}
	tokenData, err := dcdtconvert.GetFullMockDCDTData(account.Storage, systemAccStorage)
	if err != nil {
		return nil, err
	}
	var dcdtNames []string
	for dcdtName := range tokenData {
		dcdtNames = append(dcdtNames, dcdtName)
	}
	sort.Strings(dcdtNames)
	var scenDCDT []*scenmodel.DCDTData
	for _, dcdtName := range dcdtNames {
		dcdtObj := tokenData[dcdtName]

		var scenRoles []string
		for _, mockRoles := range dcdtObj.Roles {
			scenRoles = append(scenRoles, string(mockRoles))
		}

		var scenInstances []*scenmodel.DCDTInstance
		for _, mockInstance := range dcdtObj.Instances {
			var creator scenmodel.JSONBytesFromString
			if len(mockInstance.TokenMetaData.Creator) > 0 {
				creator = scenmodel.JSONBytesFromString{
					Value:    mockInstance.TokenMetaData.Creator,
					Original: ae.exprReconstructor.Reconstruct(mockInstance.TokenMetaData.Creator, er.AddressHint),
				}
			}

			var royalties scenmodel.JSONUint64
			if mockInstance.TokenMetaData.Royalties > 0 {
				royalties = scenmodel.JSONUint64{
					Value:    uint64(mockInstance.TokenMetaData.Royalties),
					Original: ae.exprReconstructor.ReconstructFromUint64(uint64(mockInstance.TokenMetaData.Royalties)),
				}
			}

			var hash scenmodel.JSONBytesFromString
			if len(mockInstance.TokenMetaData.Hash) > 0 {
				hash = scenmodel.JSONBytesFromString{
					Value:    mockInstance.TokenMetaData.Hash,
					Original: ae.exprReconstructor.Reconstruct(mockInstance.TokenMetaData.Hash, er.NoHint),
				}
			}

			var jsonUris []scenmodel.JSONBytesFromString
			for _, uri := range mockInstance.TokenMetaData.URIs {
				jsonUris = append(jsonUris, scenmodel.JSONBytesFromString{
					Value:    uri,
					Original: ae.exprReconstructor.Reconstruct(uri, er.StrHint),
				})
			}

			var attributes scenmodel.JSONBytesFromTree
			if len(mockInstance.TokenMetaData.Attributes) > 0 {
				attributes = scenmodel.JSONBytesFromTree{
					Value:    mockInstance.TokenMetaData.Attributes,
					Original: &oj.OJsonString{Value: ae.exprReconstructor.Reconstruct(mockInstance.TokenMetaData.Attributes, er.NoHint)},
				}
			}

			scenInstances = append(scenInstances, &scenmodel.DCDTInstance{
				Nonce: scenmodel.JSONUint64{
					Value:    mockInstance.TokenMetaData.Nonce,
					Original: ae.exprReconstructor.ReconstructFromUint64(mockInstance.TokenMetaData.Nonce),
				},
				Balance: scenmodel.JSONBigInt{
					Value:    mockInstance.Value,
					Original: ae.exprReconstructor.ReconstructFromBigInt(mockInstance.Value),
				},
				Creator:    creator,
				Royalties:  royalties,
				Hash:       hash,
				Uris:       scenmodel.JSONValueList{Values: jsonUris},
				Attributes: attributes,
			})
		}

		scenDCDT = append(scenDCDT, &scenmodel.DCDTData{
			TokenIdentifier: scenmodel.JSONBytesFromString{
				Value:    dcdtObj.TokenIdentifier,
				Original: ae.exprReconstructor.Reconstruct(dcdtObj.TokenIdentifier, er.StrHint),
			},
			Instances: scenInstances,
			LastNonce: scenmodel.JSONUint64{
				Value:    dcdtObj.LastNonce,
				Original: ae.exprReconstructor.ReconstructFromUint64(dcdtObj.LastNonce),
			},
			Roles: scenRoles,
		})
	}

	return &scenmodel.Account{
		Address: scenmodel.JSONBytesFromString{
			Value:    account.Address,
			Original: ae.exprReconstructor.Reconstruct(account.Address, er.AddressHint),
		},
		Nonce: scenmodel.JSONUint64{
			Value:    account.Nonce,
			Original: ae.exprReconstructor.ReconstructFromUint64(account.Nonce),
		},
		Balance: scenmodel.JSONBigInt{
			Value:    account.Balance,
			Original: ae.exprReconstructor.ReconstructFromBigInt(account.Balance),
		},
		Storage:  storageKvps,
		DCDTData: scenDCDT,
		Owner: scenmodel.JSONBytesFromString{
			Value:    account.OwnerAddress,
			Original: ae.exprReconstructor.Reconstruct(account.OwnerAddress, er.AddressHint),
		},
	}, nil
}

// DumpWorld prints the state of the MockWorld to stdout.
func (ae *ScenarioExecutor) DumpWorld() error {
	fmt.Print("world state dump:\n")
	var scenAccounts []*scenmodel.Account

	for _, account := range ae.World.AcctMap {
		scenAccount, err := ae.convertMockAccountToScenarioFormat(account)
		if err != nil {
			return err
		}
		scenAccounts = append(scenAccounts, scenAccount)
	}

	ojAccount := scenjwrite.AccountsToOJ(scenAccounts)
	s := oj.JSONString(ojAccount)
	fmt.Println(s)

	return nil
}
