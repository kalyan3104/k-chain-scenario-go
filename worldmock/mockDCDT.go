package worldmock

import (
	"fmt"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/core/check"
	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
	"github.com/kalyan3104/k-chain-core-go/data/vm"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
	"github.com/kalyan3104/k-chain-scenario-go/worldmock/dcdtconvert"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
)

// GetTokenBalance returns the DCDT balance of an account for the given token
// key (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) GetTokenBalance(address []byte, tokenIdentifier []byte, nonce uint64) (*big.Int, error) {
	account := bf.World.AcctMap.GetAccount(address)
	if check.IfNil(account) {
		return big.NewInt(0), nil
	}
	return dcdtconvert.GetTokenBalance(tokenIdentifier, nonce, account.Storage)
}

// GetTokenData gets the DCDT information related to a token from the storage of an account
// (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) GetTokenData(address []byte, tokenIdentifier []byte, nonce uint64) (*dcdt.DCDigitalToken, error) {
	account := bf.World.AcctMap.GetAccount(address)
	if check.IfNil(account) {
		return &dcdt.DCDigitalToken{
			Value: big.NewInt(0),
		}, nil
	}
	systemAccStorage := make(map[string][]byte)
	systemAcc := bf.World.AcctMap.GetAccount(vmcommon.SystemAccountAddress)
	if systemAcc != nil {
		systemAccStorage = systemAcc.Storage
	}
	return account.GetTokenData(tokenIdentifier, nonce, systemAccStorage)
}

// SetTokenData sets the DCDT information related to a token from the storage of an account
// (token keys are built from the token identifier using MakeTokenKey).
func (bf *BuiltinFunctionsWrapper) SetTokenData(address []byte, tokenIdentifier []byte, nonce uint64, tokenData *dcdt.DCDigitalToken) error {
	account := bf.World.AcctMap.GetAccount(address)
	if check.IfNil(account) {
		return nil
	}
	return account.SetTokenData(tokenIdentifier, nonce, tokenData)
}

// PerformDirectDCDTTransfer calls the real DCDTTransfer function immediately;
// only works for in-shard transfers for now, but it will be expanded to
// cross-shard.
// TODO rewrite to simulate what the SCProcessor does when executing a tx with
// data "DCDTTransfer@token@value@contractfunc@contractargs..."
// TODO this function duplicates code from host.ExecuteDCDTTransfer(), must refactor
func (bf *BuiltinFunctionsWrapper) PerformDirectDCDTTransfer(
	sender []byte,
	receiver []byte,
	token []byte,
	nonce uint64,
	value *big.Int,
	callType vm.CallType,
	gasLimit uint64,
	gasPrice uint64,
) (uint64, error) {
	dcdtTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    callType,
			GasPrice:    gasPrice,
			GasProvided: gasLimit,
			GasLocked:   0,
		},
		RecipientAddr:     receiver,
		Function:          core.BuiltInFunctionDCDTTransfer,
		AllowInitFunction: false,
	}

	if nonce > 0 {
		dcdtTransferInput.Function = core.BuiltInFunctionDCDTNFTTransfer
		dcdtTransferInput.RecipientAddr = dcdtTransferInput.CallerAddr
		nonceAsBytes := big.NewInt(0).SetUint64(nonce).Bytes()
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, token, nonceAsBytes, value.Bytes(), receiver)
	} else {
		dcdtTransferInput.Arguments = append(dcdtTransferInput.Arguments, token, value.Bytes())
	}

	vmOutput, err := bf.ProcessBuiltInFunction(dcdtTransferInput)
	if err != nil {
		return 0, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return 0, fmt.Errorf(
			"DCDTtransfer failed: retcode = %d, msg = %s",
			vmOutput.ReturnCode,
			vmOutput.ReturnMessage)
	}

	return vmOutput.GasRemaining, nil
}

// PerformDirectMultiDCDTTransfer -
func (bf *BuiltinFunctionsWrapper) PerformDirectMultiDCDTTransfer(
	sender []byte,
	receiver []byte,
	dcdtTransfers []*scenmodel.DCDTTxData,
	callType vm.CallType,
	gasLimit uint64,
	gasPrice uint64,
) (uint64, error) {
	nrTransfers := len(dcdtTransfers)
	nrTransfersAsBytes := big.NewInt(0).SetUint64(uint64(nrTransfers)).Bytes()

	multiTransferInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			Arguments:   make([][]byte, 0),
			CallValue:   big.NewInt(0),
			CallType:    callType,
			GasPrice:    gasPrice,
			GasProvided: gasLimit,
			GasLocked:   0,
		},
		RecipientAddr:     sender,
		Function:          core.BuiltInFunctionMultiDCDTNFTTransfer,
		AllowInitFunction: false,
	}
	multiTransferInput.Arguments = append(multiTransferInput.Arguments, receiver, nrTransfersAsBytes)

	for i := 0; i < nrTransfers; i++ {
		token := dcdtTransfers[i].TokenIdentifier.Value
		nonceAsBytes := big.NewInt(0).SetUint64(dcdtTransfers[i].Nonce.Value).Bytes()
		value := dcdtTransfers[i].Value.Value

		multiTransferInput.Arguments = append(multiTransferInput.Arguments, token, nonceAsBytes, value.Bytes())
	}

	vmOutput, err := bf.ProcessBuiltInFunction(multiTransferInput)
	if err != nil {
		return 0, err
	}

	if vmOutput.ReturnCode != vmcommon.Ok {
		return 0, fmt.Errorf(
			"MultiDCDTtransfer failed: retcode = %d, msg = %s",
			vmOutput.ReturnCode,
			vmOutput.ReturnMessage)
	}

	return vmOutput.GasRemaining, nil
}
