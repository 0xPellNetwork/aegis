package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultGenesis returns the default xmsg genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		OutTxTrackerList:   []OutTxTracker{},
		InTxHashToXmsgList: []InTxHashToXmsg{},
		GasPriceList:       []*GasPrice{},
		//Xmsg:            []*Send{},

	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in outTxTracker
	outTxTrackerIndexMap := make(map[string]struct{})

	for _, elem := range gs.OutTxTrackerList {
		index := string(OutTxTrackerKey(elem.Index))
		if _, ok := outTxTrackerIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for outTxTracker")
		}
		outTxTrackerIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in inTxHashToXmsg
	inTxHashToXmsgIndicesMap := make(map[string]struct{})

	for _, elem := range gs.InTxHashToXmsgList {
		index := string(InTxHashToXmsgKey(elem.InTxHash))
		if _, ok := inTxHashToXmsgIndicesMap[index]; ok {
			return fmt.Errorf("duplicated index for inTxHashToXmsg")
		}
		inTxHashToXmsgIndicesMap[index] = struct{}{}
	}

	// Check for duplicated index in gasPrice
	gasPriceIndexMap := make(map[string]bool)

	for _, elem := range gs.GasPriceList {
		if _, ok := gasPriceIndexMap[elem.Index]; ok {
			return fmt.Errorf("duplicated index for gasPrice")
		}
		gasPriceIndexMap[elem.Index] = true
	}

	return gs.RateLimiterFlags.Validate()
}

func GetGenesisStateFromAppState(marshaler codec.JSONCodec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		err := marshaler.UnmarshalJSON(appState[ModuleName], &genesisState)
		if err != nil {
			panic(fmt.Sprintf("Failed to get genesis state from app state: %s", err.Error()))
		}
	}
	return genesisState
}
