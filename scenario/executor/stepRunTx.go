package scenexec

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/vm"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
	vmcommon "github.com/kalyan3104/k-chain-vm-common-go"
)

// ExecuteTxStep executes a TxStep.
func (ae *ScenarioExecutor) ExecuteTxStep(step *scenmodel.TxStep) (*vmcommon.VMOutput, error) {
	log.Trace("ExecuteTxStep", "id", step.TxIdent)
	if len(step.Comment) > 0 {
		log.Trace("ExecuteTxStep", "comment", step.Comment)
	}

	if step.DisplayLogs {
		SetLoggingForTests()
	}

	output, err := ae.executeTx(step.TxIdent, step.Tx)
	if err != nil {
		return nil, err
	}

	if step.DisplayLogs {
		DisableLoggingForTests()
	}

	// check results
	if step.ExpectedResult != nil {
		err = ae.checkTxResults(step.TxIdent, step.ExpectedResult, ae.checkGas, output)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func (ae *ScenarioExecutor) executeTx(txIndex string, tx *scenmodel.Transaction) (*vmcommon.VMOutput, error) {
	ae.World.CreateStateBackup()

	var err error
	defer func() {
		if err != nil {
			errRollback := ae.World.RollbackChanges()
			if errRollback != nil {
				err = errRollback
			}
		} else {
			errCommit := ae.World.CommitChanges()
			if errCommit != nil {
				err = errCommit
			}
		}
	}()

	gasForExecution := uint64(0)

	if tx.Type.HasSender() {
		beforeErr := ae.World.UpdateWorldStateBefore(
			tx.From.Value,
			tx.GasLimit.Value,
			tx.GasPrice.Value)
		if beforeErr != nil {
			err = fmt.Errorf("could not set up tx %s: %w", txIndex, beforeErr)
			return nil, err
		}

		gasForExecution = tx.GasLimit.Value
		if tx.DCDTValue != nil {
			gasRemaining, err := ae.directDCDTTransferFromTx(tx)
			if err != nil {
				return nil, err
			}

			gasForExecution = gasRemaining
		}
	}

	// we also use fake vm outputs for transactions that don't use the VM, just for convenience
	var output *vmcommon.VMOutput

	if !ae.senderHasEnoughBalance(tx) {
		// out of funds is handled by the protocol, so it needs to be mocked here
		output = outOfFundsResult()
	} else {
		switch tx.Type {
		case scenmodel.ScDeploy:
			output, err = ae.scCreate(txIndex, tx, gasForExecution)
			if err != nil {
				return nil, err
			}
			if ae.PeekTraceGas() {
				fmt.Println("\nIn txID:", txIndex, ", step type:Deploy", ", total gas used:", gasForExecution-output.GasRemaining)
			}
		case scenmodel.ScQuery:
			// imitates the behaviour of the protocol
			// the sender is the contract itself during SC queries
			tx.From = tx.To
			// gas restrictions waived during SC queries
			tx.GasLimit.Value = math.MaxInt64
			gasForExecution = math.MaxInt64
			fallthrough
		case scenmodel.ScCall:
			output, err = ae.scCall(txIndex, tx, gasForExecution)
			if err != nil {
				return nil, err
			}
			if ae.PeekTraceGas() {
				fmt.Println("\nIn txID:", txIndex, ", step type:ScCall, function:", tx.Function, ", total gas used:", gasForExecution-output.GasRemaining)
			}
		case scenmodel.Transfer:
			output = ae.simpleTransferOutput(tx)
		case scenmodel.ValidatorReward:
			output, err = ae.validatorRewardOutput(tx)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unknown transaction type")
		}
	}

	if output.ReturnCode == vmcommon.Ok {
		err := ae.updateStateAfterTx(tx, output)
		if err != nil {
			return nil, err
		}
	} else {
		err = fmt.Errorf(
			"tx step failed: retcode=%d, msg=%s",
			output.ReturnCode, output.ReturnMessage)
	}

	return output, nil
}

func (ae *ScenarioExecutor) senderHasEnoughBalance(tx *scenmodel.Transaction) bool {
	if !tx.Type.HasSender() {
		return true
	}
	sender := ae.World.AcctMap.GetAccount(tx.From.Value)
	return sender.Balance.Cmp(tx.REWAValue.Value) >= 0
}

func (ae *ScenarioExecutor) simpleTransferOutput(tx *scenmodel.Transaction) *vmcommon.VMOutput {
	outputAccounts := make(map[string]*vmcommon.OutputAccount)
	outputAccounts[string(tx.To.Value)] = &vmcommon.OutputAccount{
		Address:      tx.To.Value,
		BalanceDelta: tx.REWAValue.Value,
	}

	return &vmcommon.VMOutput{
		ReturnData:      make([][]byte, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnMessage:   "",
		GasRemaining:    0,
		GasRefund:       big.NewInt(0),
		OutputAccounts:  outputAccounts,
		DeletedAccounts: make([][]byte, 0),
		TouchedAccounts: make([][]byte, 0),
		Logs:            make([]*vmcommon.LogEntry, 0),
	}
}

func (ae *ScenarioExecutor) validatorRewardOutput(tx *scenmodel.Transaction) (*vmcommon.VMOutput, error) {
	reward := tx.REWAValue.Value
	recipient := ae.World.AcctMap.GetAccount(tx.To.Value)
	if recipient == nil {
		return nil, fmt.Errorf("tx recipient (address: %s) does not exist", hex.EncodeToString(tx.To.Value))
	}
	recipient.BalanceDelta = reward
	storageReward := recipient.StorageValue(RewardKey)
	storageReward = big.NewInt(0).Add(
		big.NewInt(0).SetBytes(storageReward),
		reward).Bytes()

	outputAccounts := make(map[string]*vmcommon.OutputAccount)
	outputAccounts[string(tx.To.Value)] = &vmcommon.OutputAccount{
		Address:      tx.To.Value,
		BalanceDelta: tx.REWAValue.Value,
		StorageUpdates: map[string]*vmcommon.StorageUpdate{
			RewardKey: {
				Offset: []byte(RewardKey),
				Data:   storageReward,
			},
		},
	}

	return &vmcommon.VMOutput{
		ReturnData:      make([][]byte, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnMessage:   "",
		GasRemaining:    0,
		GasRefund:       big.NewInt(0),
		OutputAccounts:  outputAccounts,
		DeletedAccounts: make([][]byte, 0),
		TouchedAccounts: make([][]byte, 0),
		Logs:            make([]*vmcommon.LogEntry, 0),
	}, nil
}

func outOfFundsResult() *vmcommon.VMOutput {
	return &vmcommon.VMOutput{
		ReturnData:      make([][]byte, 0),
		ReturnCode:      vmcommon.OutOfFunds,
		ReturnMessage:   "",
		GasRemaining:    0,
		GasRefund:       big.NewInt(0),
		OutputAccounts:  make(map[string]*vmcommon.OutputAccount),
		DeletedAccounts: make([][]byte, 0),
		TouchedAccounts: make([][]byte, 0),
		Logs:            make([]*vmcommon.LogEntry, 0),
	}
}

func (ae *ScenarioExecutor) scCreate(txIndex string, tx *scenmodel.Transaction, gasLimit uint64) (*vmcommon.VMOutput, error) {
	txHash := generateTxHash(txIndex)
	vmInput := vmcommon.VMInput{
		CallerAddr:     tx.From.Value,
		Arguments:      scenmodel.JSONBytesFromTreeValues(tx.Arguments),
		CallValue:      tx.REWAValue.Value,
		GasPrice:       tx.GasPrice.Value,
		GasProvided:    gasLimit,
		OriginalTxHash: txHash,
		CurrentTxHash:  txHash,
		DCDTTransfers:  make([]*vmcommon.DCDTTransfer, 0),
	}
	addDCDTToVMInput(tx.DCDTValue, &vmInput)
	codeMetadata := tx.CodeMetadata.Value
	if tx.CodeMetadata.Unspecified {
		codeMetadata = DefaultCodeMetadata
	}
	input := &vmcommon.ContractCreateInput{
		ContractCode:         tx.Code.Value,
		ContractCodeMetadata: codeMetadata,
		VMInput:              vmInput,
	}

	return ae.vm.RunSmartContractCreate(input)
}

func (ae *ScenarioExecutor) scCall(txIndex string, tx *scenmodel.Transaction, gasLimit uint64) (*vmcommon.VMOutput, error) {
	recipient := ae.World.AcctMap.GetAccount(tx.To.Value)
	if recipient == nil {
		return nil, fmt.Errorf("tx recipient (address: %s) does not exist", hex.EncodeToString(tx.To.Value))
	}
	if len(recipient.Code) == 0 {
		return nil, fmt.Errorf("tx recipient (address: %s) is not a smart contract", hex.EncodeToString(tx.To.Value))
	}
	txHash := generateTxHash(txIndex)
	vmInput := vmcommon.VMInput{
		CallerAddr:     tx.From.Value,
		Arguments:      scenmodel.JSONBytesFromTreeValues(tx.Arguments),
		CallValue:      tx.REWAValue.Value,
		GasPrice:       tx.GasPrice.Value,
		GasProvided:    gasLimit,
		OriginalTxHash: txHash,
		CurrentTxHash:  txHash,
		DCDTTransfers:  make([]*vmcommon.DCDTTransfer, 0),
	}
	addDCDTToVMInput(tx.DCDTValue, &vmInput)
	input := &vmcommon.ContractCallInput{
		RecipientAddr: tx.To.Value,
		Function:      tx.Function,
		VMInput:       vmInput,
	}

	return ae.vm.RunSmartContractCall(input)
}

func (ae *ScenarioExecutor) directDCDTTransferFromTx(tx *scenmodel.Transaction) (uint64, error) {
	nrTransfers := len(tx.DCDTValue)

	if nrTransfers == 1 {
		return ae.World.BuiltinFuncs.PerformDirectDCDTTransfer(
			tx.From.Value,
			tx.To.Value,
			tx.DCDTValue[0].TokenIdentifier.Value,
			tx.DCDTValue[0].Nonce.Value,
			tx.DCDTValue[0].Value.Value,
			vm.DirectCall,
			tx.GasLimit.Value,
			tx.GasPrice.Value)
	} else {
		return ae.World.BuiltinFuncs.PerformDirectMultiDCDTTransfer(
			tx.From.Value,
			tx.To.Value,
			tx.DCDTValue,
			vm.DirectCall,
			tx.GasLimit.Value,
			tx.GasPrice.Value)
	}
}

func (ae *ScenarioExecutor) updateStateAfterTx(
	tx *scenmodel.Transaction,
	output *vmcommon.VMOutput) error {

	// subtract call value from sender (this is not reflected in the delta)
	// except for validatorReward, there is no sender there
	if tx.Type.HasSender() {
		_ = ae.World.UpdateBalanceWithDelta(tx.From.Value, big.NewInt(0).Neg(tx.REWAValue.Value))
	}

	// update accounts based on deltas
	updErr := ae.World.UpdateAccounts(output.OutputAccounts, output.DeletedAccounts)
	if updErr != nil {
		return updErr
	}

	// sum of all balance deltas should equal call value (unless we got an error)
	// (unless it is validatorReward, when funds just pop into existence)
	if tx.Type.HasSender() {
		sumOfBalanceDeltas := big.NewInt(0)
		for _, oa := range output.OutputAccounts {
			sumOfBalanceDeltas = sumOfBalanceDeltas.Add(sumOfBalanceDeltas, oa.BalanceDelta)
		}
		if sumOfBalanceDeltas.Cmp(tx.REWAValue.Value) != 0 {
			return fmt.Errorf("sum of balance deltas should equal call value. Sum of balance deltas: %d (0x%x). Call value: %d (0x%x)",
				sumOfBalanceDeltas, sumOfBalanceDeltas, tx.REWAValue.Value, tx.REWAValue.Value)
		}
	}

	return nil
}

func generateTxHash(txIndex string) []byte {
	txIndexBytes := []byte(txIndex)
	if len(txIndexBytes) > 32 {
		return txIndexBytes[:32]
	}
	for i := len(txIndexBytes); i < 32; i++ {
		txIndexBytes = append(txIndexBytes, '.')
	}
	return txIndexBytes
}

func addDCDTToVMInput(dcdtData []*scenmodel.DCDTTxData, vmInput *vmcommon.VMInput) {
	dcdtDataLen := len(dcdtData)

	if dcdtDataLen > 0 {
		vmInput.DCDTTransfers = make([]*vmcommon.DCDTTransfer, dcdtDataLen)
		for i := 0; i < dcdtDataLen; i++ {
			vmInput.DCDTTransfers[i] = &vmcommon.DCDTTransfer{}
			vmInput.DCDTTransfers[i].DCDTTokenName = dcdtData[i].TokenIdentifier.Value
			vmInput.DCDTTransfers[i].DCDTValue = dcdtData[i].Value.Value
			vmInput.DCDTTransfers[i].DCDTTokenNonce = dcdtData[i].Nonce.Value
			if vmInput.DCDTTransfers[i].DCDTTokenNonce != 0 {
				vmInput.DCDTTransfers[i].DCDTTokenType = uint32(core.NonFungible)
			} else {
				vmInput.DCDTTransfers[i].DCDTTokenType = uint32(core.Fungible)
			}
		}
	}
}