package scenjsonparse

import (
	"errors"
	"fmt"

	oj "github.com/kalyan3104/k-chain-scenario-go/orderedjson"
	scenmodel "github.com/kalyan3104/k-chain-scenario-go/scenario/model"
)

func (p *Parser) processTxDCDT(txDcdtRaw oj.OJsonObject) ([]*scenmodel.DCDTTxData, error) {
	allDcdtData := make([]*scenmodel.DCDTTxData, 0)

	switch txDcdt := txDcdtRaw.(type) {
	case *oj.OJsonMap:
		if !p.AllowDcdtTxLegacySyntax {
			return nil, fmt.Errorf("wrong DCDT Multi-Transfer format, list expected")
		}
		entry, err := p.parseSingleTxDcdtEntry(txDcdt)
		if err != nil {
			return nil, err
		}

		allDcdtData = append(allDcdtData, entry)
	case *oj.OJsonList:
		for _, txDcdtListItem := range txDcdt.AsList() {
			txDcdtMap, isMap := txDcdtListItem.(*oj.OJsonMap)
			if !isMap {
				return nil, fmt.Errorf("wrong DCDT Multi-Transfer format")
			}

			entry, err := p.parseSingleTxDcdtEntry(txDcdtMap)
			if err != nil {
				return nil, err
			}

			allDcdtData = append(allDcdtData, entry)
		}
	default:
		return nil, fmt.Errorf("wrong DCDT transfer format, expected list")
	}

	return allDcdtData, nil
}

func (p *Parser) parseSingleTxDcdtEntry(dcdtTxEntry *oj.OJsonMap) (*scenmodel.DCDTTxData, error) {
	dcdtData := scenmodel.DCDTTxData{}
	var err error

	for _, kvp := range dcdtTxEntry.OrderedKV {
		switch kvp.Key {
		case "tokenIdentifier":
			dcdtData.TokenIdentifier, err = p.processStringAsByteArray(kvp.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid DCDT token name: %w", err)
			}
		case "nonce":
			dcdtData.Nonce, err = p.processUint64(kvp.Value)
			if err != nil {
				return nil, errors.New("invalid account nonce")
			}
		case "value":
			dcdtData.Value, err = p.processBigInt(kvp.Value, bigIntUnsignedBytes)
			if err != nil {
				return nil, fmt.Errorf("invalid DCDT balance: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown transaction DCDT data field: %s", kvp.Key)
		}
	}

	return &dcdtData, nil
}
