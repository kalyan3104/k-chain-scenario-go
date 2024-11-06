package scenjsonparse

import (
	"errors"
	"fmt"

	oj "github.com/kalyan3104/k-chain-scenario-go/orderedjson"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
)

func (p *Parser) processCheckDCDTData(
	tokenName scenmodel.JSONBytesFromString,
	dcdtDataRaw oj.OJsonObject) (*scenmodel.CheckDCDTData, error) {

	switch data := dcdtDataRaw.(type) {
	case *oj.OJsonString:
		// simple string representing balance "400,000,000,000"
		dcdtData := scenmodel.CheckDCDTData{
			TokenIdentifier: tokenName,
		}
		balance, err := p.processCheckBigInt(dcdtDataRaw, bigIntUnsignedBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid DCDT balance: %w", err)
		}
		dcdtData.Instances = []*scenmodel.CheckDCDTInstance{
			{
				Nonce:   scenmodel.JSONUint64Zero(),
				Balance: balance,
			},
		}
		return &dcdtData, nil
	case *oj.OJsonMap:
		return p.processCheckDCDTDataMap(tokenName, data)
	default:
		return nil, errors.New("invalid JSON object for DCDT")
	}
}

// Map containing DCDT fields, e.g.:
//
//	{
//		"instances": [ ... ],
//	 "lastNonce": "5",
//		"frozen": "true"
//	}
func (p *Parser) processCheckDCDTDataMap(tokenName scenmodel.JSONBytesFromString, dcdtDataMap *oj.OJsonMap) (*scenmodel.CheckDCDTData, error) {
	dcdtData := scenmodel.CheckDCDTData{
		TokenIdentifier: tokenName,
	}
	// var err error
	firstInstance := &scenmodel.CheckDCDTInstance{
		Nonce:      scenmodel.JSONUint64Zero(),
		Balance:    scenmodel.JSONCheckBigIntUnspecified(),
		Creator:    scenmodel.JSONCheckBytesUnspecified(),
		Royalties:  scenmodel.JSONCheckUint64Unspecified(),
		Hash:       scenmodel.JSONCheckBytesUnspecified(),
		Uris:       scenmodel.JSONCheckValueListUnspecified(),
		Attributes: scenmodel.JSONCheckBytesUnspecified(),
	}
	firstInstanceLoaded := false
	var explicitInstances []*scenmodel.CheckDCDTInstance

	for _, kvp := range dcdtDataMap.OrderedKV {
		// it is allowed to load the instance directly, fields set to the first instance
		instanceFieldLoaded, err := p.tryProcessCheckDCDTInstanceField(kvp, firstInstance)
		if err != nil {
			return nil, fmt.Errorf("invalid account DCDT instance field: %w", err)
		}
		if instanceFieldLoaded {
			firstInstanceLoaded = true
		} else {
			switch kvp.Key {
			case "instances":
				explicitInstances, err = p.processCheckDCDTInstances(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT instances: %w", err)
				}
			case "lastNonce":
				dcdtData.LastNonce, err = p.processCheckUint64(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT lastNonce: %w", err)
				}
			case "roles":
				dcdtData.Roles, err = p.processStringList(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid account DCDT roles: %w", err)
				}
			case "frozen":
				dcdtData.Frozen, err = p.processCheckUint64(kvp.Value)
				if err != nil {
					return nil, fmt.Errorf("invalid DCDT frozen flag: %w", err)
				}
			default:
				return nil, fmt.Errorf("unknown DCDT data field: %s", kvp.Key)
			}
		}
	}

	if firstInstanceLoaded {
		if !p.AllowDcdtLegacyCheckSyntax {
			return nil, fmt.Errorf("wrong DCDT check state syntax: instances in root no longer allowed")
		}
		dcdtData.Instances = []*scenmodel.CheckDCDTInstance{firstInstance}
	}
	dcdtData.Instances = append(dcdtData.Instances, explicitInstances...)

	return &dcdtData, nil
}

func (p *Parser) tryProcessCheckDCDTInstanceField(kvp *oj.OJsonKeyValuePair, targetInstance *scenmodel.CheckDCDTInstance) (bool, error) {
	var err error
	switch kvp.Key {
	case "nonce":
		targetInstance.Nonce, err = p.processUint64(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid account nonce: %w", err)
		}
	case "balance":
		targetInstance.Balance, err = p.processCheckBigInt(kvp.Value, bigIntUnsignedBytes)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT balance: %w", err)
		}
	case "creator":
		targetInstance.Creator, err = p.parseCheckBytes(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT creator address: %w", err)
		}
	case "royalties":
		targetInstance.Royalties, err = p.processCheckUint64(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT royalties: %w", err)
		}
		if targetInstance.Royalties.Value > 10000 {
			return false, errors.New("invalid DCDT NFT royalties: value exceeds maximum allowed 10000")
		}
	case "hash":
		targetInstance.Hash, err = p.parseCheckBytes(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT hash: %w", err)
		}
	case "uri":
		targetInstance.Uris, err = p.parseCheckValueList(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT URI: %w", err)
		}
	case "attributes":
		targetInstance.Attributes, err = p.parseCheckBytes(kvp.Value)
		if err != nil {
			return false, fmt.Errorf("invalid DCDT NFT attributes: %w", err)
		}
	default:
		return false, nil
	}
	return true, nil
}

func (p *Parser) processCheckDCDTInstances(dcdtInstancesRaw oj.OJsonObject) ([]*scenmodel.CheckDCDTInstance, error) {
	var instancesResult []*scenmodel.CheckDCDTInstance
	dcdtInstancesList, isList := dcdtInstancesRaw.(*oj.OJsonList)
	if !isList {
		return nil, errors.New("dcdt instances object is not a list")
	}
	for _, instanceItem := range dcdtInstancesList.AsList() {
		instanceAsMap, isMap := instanceItem.(*oj.OJsonMap)
		if !isMap {
			return nil, errors.New("JSON map expected as dcdt instances list item")
		}

		instance := scenmodel.NewCheckDCDTInstance()

		for _, kvp := range instanceAsMap.OrderedKV {
			instanceFieldLoaded, err := p.tryProcessCheckDCDTInstanceField(kvp, instance)
			if err != nil {
				return nil, fmt.Errorf("invalid account DCDT instance field in instances list: %w", err)
			}
			if !instanceFieldLoaded {
				return nil, fmt.Errorf("invalid account DCDT instance field in instances list: `%s`", kvp.Key)
			}
		}

		instancesResult = append(instancesResult, instance)

	}

	return instancesResult, nil
}
