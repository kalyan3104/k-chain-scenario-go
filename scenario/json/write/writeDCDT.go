package scenjsonwrite

import (
	oj "github.com/kalyan3104/k-chain-scenario-go/orderedjson"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
)

func dcdtTxDataToOJ(dcdtItems []*scenmodel.DCDTTxData) oj.OJsonObject {
	dcdtItemList := oj.OJsonList{}
	for _, dcdtItemRaw := range dcdtItems {
		dcdtItemOJ := dcdtTxRawEntryToOJ(dcdtItemRaw)
		dcdtItemList = append(dcdtItemList, dcdtItemOJ)
	}

	return &dcdtItemList

}

func dcdtTxRawEntryToOJ(dcdtItemRaw *scenmodel.DCDTTxData) *oj.OJsonMap {
	dcdtItemOJ := oj.NewMap()

	if len(dcdtItemRaw.TokenIdentifier.Original) > 0 {
		dcdtItemOJ.Put("tokenIdentifier", bytesFromStringToOJ(dcdtItemRaw.TokenIdentifier))
	}
	if len(dcdtItemRaw.Nonce.Original) > 0 {
		dcdtItemOJ.Put("nonce", uint64ToOJ(dcdtItemRaw.Nonce))
	}
	if len(dcdtItemRaw.Value.Original) > 0 {
		dcdtItemOJ.Put("value", bigIntToOJ(dcdtItemRaw.Value))
	}

	return dcdtItemOJ
}

func dcdtDataToOJ(dcdtItems []*scenmodel.DCDTData) *oj.OJsonMap {
	dcdtItemsOJ := oj.NewMap()
	for _, dcdtItem := range dcdtItems {
		dcdtItemsOJ.Put(dcdtItem.TokenIdentifier.Original, dcdtItemToOJ(dcdtItem))
	}
	return dcdtItemsOJ
}

func dcdtItemToOJ(dcdtItem *scenmodel.DCDTData) oj.OJsonObject {
	if isCompactDCDT(dcdtItem) {
		return bigIntToOJ(dcdtItem.Instances[0].Balance)
	}

	dcdtItemOJ := oj.NewMap()

	// instances
	if len(dcdtItem.Instances) > 0 {
		var convertedList []oj.OJsonObject
		for _, dcdtInstance := range dcdtItem.Instances {
			dcdtInstanceOJ := oj.NewMap()
			appendDCDTInstanceToOJ(dcdtInstance, dcdtInstanceOJ)
			convertedList = append(convertedList, dcdtInstanceOJ)
		}
		instancesOJList := oj.OJsonList(convertedList)
		dcdtItemOJ.Put("instances", &instancesOJList)
	}

	if len(dcdtItem.LastNonce.Original) > 0 {
		dcdtItemOJ.Put("lastNonce", uint64ToOJ(dcdtItem.LastNonce))
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
		dcdtItemOJ.Put("frozen", uint64ToOJ(dcdtItem.Frozen))
	}

	return dcdtItemOJ
}

func appendDCDTInstanceToOJ(dcdtInstance *scenmodel.DCDTInstance, targetOj *oj.OJsonMap) {
	targetOj.Put("nonce", uint64ToOJ(dcdtInstance.Nonce))

	if len(dcdtInstance.Balance.Original) > 0 {
		targetOj.Put("balance", bigIntToOJ(dcdtInstance.Balance))
	}
	if len(dcdtInstance.Creator.Original) > 0 {
		targetOj.Put("creator", bytesFromStringToOJ(dcdtInstance.Creator))
	}
	if len(dcdtInstance.Royalties.Original) > 0 {
		targetOj.Put("royalties", uint64ToOJ(dcdtInstance.Royalties))
	}
	if len(dcdtInstance.Hash.Original) > 0 {
		targetOj.Put("hash", bytesFromStringToOJ(dcdtInstance.Hash))
	}
	if !dcdtInstance.Uris.IsUnspecified() {
		targetOj.Put("uri", valueListToOJ(dcdtInstance.Uris))
	}
	if len(dcdtInstance.Attributes.Value) > 0 {
		targetOj.Put("attributes", bytesFromTreeToOJ(dcdtInstance.Attributes))
	}
}

func isCompactDCDT(dcdtItem *scenmodel.DCDTData) bool {
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
