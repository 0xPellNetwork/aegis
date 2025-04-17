package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/0xPellNetwork/aegis/app"
)

func TestParsefileToObserverMapper(t *testing.T) {
	file := "tmp.json"
	defer func(t *testing.T, fp string) {
		err := os.RemoveAll(fp)
		require.NoError(t, err)
	}(t, file)
	app.SetConfig()
	createRelayerList(file)
	obsListReadFromFile, err := ParsefileToObserverDetails(file)
	require.NoError(t, err)
	for _, obs := range obsListReadFromFile {
		require.Equal(t, obs.PellClientGranteeAddress, sdk.AccAddress(crypto.AddressHash([]byte("ObserverGranteeAddress"))).String())
	}
}

func createRelayerList(fp string) {
	var listReader []ObserverInfoReader
	//listChainID := []int64{common.GoerliLocalNetChain().ChainId, common.BtcRegtestChain().ChainId, common.PellChain().ChainId}
	commonGrantAddress := sdk.AccAddress(crypto.AddressHash([]byte("ObserverGranteeAddress")))
	observerAddress := sdk.AccAddress(crypto.AddressHash([]byte("ObserverAddress")))
	validatorAddress := sdk.ValAddress(crypto.AddressHash([]byte("ValidatorAddress")))
	info := ObserverInfoReader{
		ObserverAddress:           observerAddress.String(),
		PellClientGranteeAddress:  commonGrantAddress.String(),
		StakingGranteeAddress:     commonGrantAddress.String(),
		StakingMaxTokens:          "100000000",
		StakingValidatorAllowList: []string{validatorAddress.String()},
		SpendMaxTokens:            "100000000",
		GovGranteeAddress:         commonGrantAddress.String(),
		PellClientGranteePubKey:   "pellpub1addwnpepqggtjvkmj6apcqr6ynyc5edxf2mpf5fxp2d3kwupemxtfwvg6gm7qv79fw0",
	}
	listReader = append(listReader, info)

	file, _ := json.MarshalIndent(listReader, "", " ")
	_ = ioutil.WriteFile(fp, file, 0600)
}
