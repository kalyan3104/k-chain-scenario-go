package scenjsonwrite

import (
	oj "github.com/kalyan3104/k-chain-scenario-go/orderedjson"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
)

func checkDCDTDataToOJ(dcdtItems []*scenmodel.CheckDCDTData, moreDCDTTokensAllowed bool) *oj.OJsonMap {
	dcdtItemsOJ := oj.NewMap()
	for _, dcdtItem := range dcdtItems {
		dcdtItemsOJ.Put(dcdtItem.TokenIdentifier.Original, checkDCDTItemToOJ(dcdtItem))
	}
	if moreDCDTTokensAllowed {
		dcdtItemsOJ.Put("+", stringToOJ(""))
	}
	return dcdtItemsOJ
}

func checkDCDTItemToOJ(dcdtItem *scenmodel.CheckDCDTData) oj.OJsonObject {
	if isCompactCheckDCDT(dcdtItem) {
		return checkBigIntToOJ(dcdtItem.Instances[0].Balance)
	}

	dcdtItemOJ := oj.NewMap()

	// instances
	if len(dcdtItem.Instances) > 0 {
		var convertedList []oj.OJsonObject
		for _, dcdtInstance := range dcdtItem.Instances {
			dcdtInstanceOJ := oj.NewMap()
			appendCheckDCDTInstanceToOJ(dcdtInstance, dcdtInstanceOJ)
			convertedList = append(convertedList, dcdtInstanceOJ)
		}
		instancesOJList := oj.OJsonList(convertedList)
		dcdtItemOJ.Put("instances", &instancesOJList)
	}

	if len(dcdtItem.LastNonce.Original) > 0 {
		dcdtItemOJ.Put("lastNonce", checkUint64ToOJ(dcdtItem.LastNonce))
	}

	// roles
	if len(dcdtItem.Roles) > 0 {
		var convertedList []oj.OJsonObject
		for _, roleStr := range dcdtItem.Roles {
			convertedList = append(convertedList, &oj.OJsonString{Value: roleStr})
		}
		rolesOJList := oj.OJsonList(convertedList)
		dcdtItemOJ.Put("roles", &rolesOJList)
	}
	if len(dcdtItem.Frozen.Original) > 0 {
		dcdtItemOJ.Put("frozen", checkUint64ToOJ(dcdtItem.Frozen))
	}

	return dcdtItemOJ
}

func appendCheckDCDTInstanceToOJ(dcdtInstance *scenmodel.CheckDCDTInstance, targetOj *oj.OJsonMap) {
	targetOj.Put("nonce", uint64ToOJ(dcdtInstance.Nonce))

	if len(dcdtInstance.Balance.Original) > 0 {
		targetOj.Put("balance", checkBigIntToOJ(dcdtInstance.Balance))
	}
	if !dcdtInstance.Creator.Unspecified && len(dcdtInstance.Creator.Value) > 0 {
		targetOj.Put("creator", checkBytesToOJ(dcdtInstance.Creator))
	}
	if !dcdtInstance.Royalties.Unspecified && len(dcdtInstance.Royalties.Original) > 0 {
		targetOj.Put("royalties", checkUint64ToOJ(dcdtInstance.Royalties))
	}
	if !dcdtInstance.Hash.Unspecified && len(dcdtInstance.Hash.Value) > 0 {
		targetOj.Put("hash", checkBytesToOJ(dcdtInstance.Hash))
	}
	if !dcdtInstance.Uris.IsUnspecified() {
		targetOj.Put("uri", checkValueListToOJ(dcdtInstance.Uris))
	}
	if !dcdtInstance.Attributes.Unspecified && len(dcdtInstance.Attributes.Value) > 0 {
		targetOj.Put("attributes", checkBytesToOJ(dcdtInstance.Attributes))
	}
}

func isCompactCheckDCDT(dcdtItem *scenmodel.CheckDCDTData) bool {
	if len(dcdtItem.Instances) != 1 {
		return false
	}
	if len(dcdtItem.Instances[0].Nonce.Original) > 0 {
		return false
	}
	if len(dcdtItem.Roles) > 0 {
		return false
	}
	if len(dcdtItem.Frozen.Original) > 0 {
		return false
	}
	return true
}
