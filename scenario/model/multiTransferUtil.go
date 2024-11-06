package scenmodel

import (
	"github.com/kalyan3104/k-chain-core-go/core"
	txDataBuilder "github.com/kalyan3104/k-chain-vm-common-go/txDataBuilder"
)

// CreateMultiTransferData builds data for a multiTransferDCDT
func CreateMultiTransferData(to []byte, dcdtData []*DCDTTxData, endpointName string, arguments [][]byte) []byte {
	multiTransferData := make([]byte, 0)
	multiTransferData = append(multiTransferData, []byte(core.BuiltInFunctionMultiDCDTNFTTransfer)...)
	tdb := txDataBuilder.NewBuilder()
	tdb.Bytes(to)
	tdb.Int(len(dcdtData))

	for _, dcdtDataTransfer := range dcdtData {
		tdb.Bytes(dcdtDataTransfer.TokenIdentifier.Value)
		tdb.Int64(int64(dcdtDataTransfer.Nonce.Value))
		tdb.BigInt(dcdtDataTransfer.Value.Value)
	}

	if len(endpointName) > 0 {
		tdb.Str(endpointName)

		for _, arg := range arguments {
			tdb.Bytes(arg)
		}
	}
	multiTransferData = append(multiTransferData, tdb.ToBytes()...)
	return multiTransferData
}
