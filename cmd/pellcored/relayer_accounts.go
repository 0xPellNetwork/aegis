package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/aegis/app"
	"github.com/0xPellNetwork/aegis/cmd/pellcored/config"
	"github.com/0xPellNetwork/aegis/pkg/crypto"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
	xmsgtypes "github.com/0xPellNetwork/aegis/x/xmsg/types"
)

// Token distribution
// Validators Only = ValidatorTokens sent to their operator address
// Observer = ObserverTokens sent to their operator address + HotkeyTokens sent to their hotkey address
// HotkeyTokens are for operational expenses such as paying for gas fees
const (
	keygenBlock = "keygen-block"
	tssPubKey   = "tss-pubkey"
)

func AddObserverAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-observer-list [observer-list.json] validator-token-amount hotkey-token-amount observer-token-amount",
		Short: "Add a list of observers to the observer mapper ,default path is ~/.pellcored/os_info/observer_info.json",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec
			serverCtx := server.GetServerContextFromCmd(cmd)
			serverConfig := serverCtx.Config

			keyGenBlock, err := cmd.Flags().GetInt64(keygenBlock)
			if err != nil {
				return err
			}
			tssPubkey, err := cmd.Flags().GetString(tssPubKey)
			if err != nil {
				return err
			}
			if keyGenBlock == 0 && tssPubkey == "" {
				panic("TSS pubkey is required if keygen block is set to 0")
			}

			path, err := url.JoinPath(app.DefaultNodeHome, args[0])
			if err != nil {
				return err
			}

			observerInfo, err := ParsefileToObserverDetails(path)
			if err != nil {
				return err
			}

			var observerSet types.RelayerSet
			var grantAuthorizations []authz.GrantAuthorization
			var nodeAccounts []*types.NodeAccount
			var keygenPubKeys []string

			var balances []banktypes.Balance
			validatorTokens, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				panic("Failed to parse string to int for observer")
			}

			hotkeyTokens, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				panic("Failed to parse string to int for hotkey")
			}

			observerTokens, ok := sdkmath.NewIntFromString(args[3])
			if !ok {
				panic("Failed to parse string to int for hotkey")
			}

			ValidatorBalance := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, validatorTokens))
			HotkeyBalance := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, hotkeyTokens))
			ObserverBalance := sdk.NewCoins(sdk.NewCoin(config.BaseDenom, observerTokens))
			// Generate the grant authorizations and created observer list for chain
			for _, info := range observerInfo {

				if isValidatorOnly(info.IsObserver) {
					balances = append(balances, banktypes.Balance{
						Address: info.ObserverAddress,
						Coins:   ValidatorBalance,
					})
					continue
				}
				balances = append(balances, banktypes.Balance{
					Address: info.ObserverAddress,
					Coins:   ValidatorBalance.Add(ObserverBalance...),
				})

				if info.PellClientGranteeAddress == "" || info.ObserverAddress == "" {
					panic("PellClientGranteeAddress or ObserverAddress is empty")
				}
				grantAuthorizations = append(grantAuthorizations, generateGrants(info)...)

				observerSet.RelayerList = append(observerSet.RelayerList, info.ObserverAddress)
				if info.PellClientGranteePubKey != "" {
					pubkey, err := crypto.NewPubKey(info.PellClientGranteePubKey)
					if err != nil {
						panic(err)
					}
					pubkeySet := crypto.PubKeySet{
						Secp256k1: pubkey,
						Ed25519:   "",
					}
					na := types.NodeAccount{
						Operator:       info.ObserverAddress,
						GranteeAddress: info.PellClientGranteeAddress,
						GranteePubkey:  &pubkeySet,
						NodeStatus:     types.NodeStatus_ACTIVE,
					}
					nodeAccounts = append(nodeAccounts, &na)
				}

				balances = append(balances, banktypes.Balance{
					Address: info.PellClientGranteeAddress,
					Coins:   HotkeyBalance,
				})
				keygenPubKeys = append(keygenPubKeys, info.PellClientGranteePubKey)
			}

			genFile := serverConfig.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// Add node accounts to cross chain genesis state
			pellCrossChainGenState := xmsgtypes.GetGenesisStateFromAppState(cdc, appState)
			tss := types.TSS{}
			if keyGenBlock == 0 {
				operatorList := make([]string, len(nodeAccounts))
				for i, nodeAccount := range nodeAccounts {
					operatorList[i] = nodeAccount.Operator
				}
				tss = types.TSS{
					TssPubkey:           tssPubkey,
					TssParticipantList:  keygenPubKeys,
					OperatorAddressList: operatorList,
					FinalizedPellHeight: 0,
					KeygenPellHeight:    0,
				}
			}
			observerSet.RelayerList = removeDuplicate(observerSet.RelayerList)
			// Add observers to observer genesis state
			pellObserverGenState := types.GetGenesisStateFromAppState(cdc, appState)
			pellObserverGenState.Observers = observerSet
			pellObserverGenState.NodeAccountList = nodeAccounts
			pellObserverGenState.Tss = &tss
			keyGenStatus := types.KeygenStatus_PENDING
			if keyGenBlock == 0 {
				keyGenStatus = types.KeygenStatus_SUCCESS
			}
			pellObserverGenState.Keygen = &types.Keygen{
				Status:         keyGenStatus,
				GranteePubkeys: keygenPubKeys,
				BlockNumber:    keyGenBlock,
			}

			// Add grant authorizations to authz genesis state
			var authzGenState authz.GenesisState
			if appState[authz.ModuleName] != nil {
				err := cdc.UnmarshalJSON(appState[authz.ModuleName], &authzGenState)
				if err != nil {
					panic(fmt.Sprintf("Failed to get genesis state from app state: %s", err.Error()))
				}
			}

			authzGenState.Authorization = grantAuthorizations

			// Marshal modified states into genesis file
			pellCrossChainStateBz, err := json.Marshal(pellCrossChainGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal Observer List into Genesis File: %w", err)
			}
			pellObserverStateBz, err := json.Marshal(pellObserverGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal Observer List into Genesis File: %w", err)
			}
			err = codectypes.UnpackInterfaces(authzGenState, cdc)
			if err != nil {
				return fmt.Errorf("failed to authz grants into upackeder: %w", err)
			}
			authZStateBz, err := cdc.MarshalJSON(&authzGenState)
			if err != nil {
				return fmt.Errorf("failed to authz grants into Genesis File: %w", err)
			}
			appState[types.ModuleName] = pellObserverStateBz
			appState[authz.ModuleName] = authZStateBz
			appState[xmsgtypes.ModuleName] = pellCrossChainStateBz
			modifiedAppState, err := AddGenesisAccount(clientCtx, balances, appState)
			if err != nil {
				panic(err)
			}
			// Create new genesis file
			appStateJSON, err := json.Marshal(modifiedAppState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}
	cmd.Flags().Int64(keygenBlock, 20, "set keygen block , default is 20")
	cmd.Flags().String(tssPubKey, "", "set TSS pubkey if using older keygen")
	return cmd
}

func removeDuplicate[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func generateGrants(info ObserverInfoReader) []authz.GrantAuthorization {
	sdk.MustAccAddressFromBech32(info.ObserverAddress)
	var grants []authz.GrantAuthorization
	if info.PellClientGranteeAddress != "" {
		sdk.MustAccAddressFromBech32(info.PellClientGranteeAddress)
		grants = append(grants, addPellClientGrants(grants, info)...)
	}
	if info.SpendGranteeAddress != "" {
		sdk.MustAccAddressFromBech32(info.SpendGranteeAddress)
		grants = append(grants, addSpendingGrants(grants, info)...)
	}
	if info.StakingGranteeAddress != "" {
		sdk.MustAccAddressFromBech32(info.StakingGranteeAddress)
		grants = append(grants, addStakingGrants(grants, info)...)
	}

	if info.GovGranteeAddress != "" {
		sdk.MustAccAddressFromBech32(info.GovGranteeAddress)
		grants = append(grants, addGovGrants(grants, info)...)
	}

	return grants
}

func addPellClientGrants(grants []authz.GrantAuthorization, info ObserverInfoReader) []authz.GrantAuthorization {
	txTypes := xmsgtypes.GetAllAuthzPellclientTxTypes()
	for _, txType := range txTypes {
		auth, err := codectypes.NewAnyWithValue(authz.NewGenericAuthorization(txType))
		if err != nil {
			panic(err)
		}
		grants = append(grants, authz.GrantAuthorization{
			Granter:       info.ObserverAddress,
			Grantee:       info.PellClientGranteeAddress,
			Authorization: auth,
			Expiration:    nil,
		})
	}
	return grants
}

func addGovGrants(grants []authz.GrantAuthorization, info ObserverInfoReader) []authz.GrantAuthorization {

	txTypes := []string{sdk.MsgTypeURL(&v1beta1.MsgVote{}),
		sdk.MsgTypeURL(&v1beta1.MsgSubmitProposal{}),
		sdk.MsgTypeURL(&v1beta1.MsgDeposit{}),
		sdk.MsgTypeURL(&v1beta1.MsgVoteWeighted{}),
		sdk.MsgTypeURL(&v1.MsgVote{}),
		sdk.MsgTypeURL(&v1.MsgSubmitProposal{}),
		sdk.MsgTypeURL(&v1.MsgDeposit{}),
		sdk.MsgTypeURL(&v1.MsgVoteWeighted{}),
	}
	for _, txType := range txTypes {
		auth, err := codectypes.NewAnyWithValue(authz.NewGenericAuthorization(txType))
		if err != nil {
			panic(err)
		}
		grants = append(grants, authz.GrantAuthorization{
			Granter:       info.ObserverAddress,
			Grantee:       info.GovGranteeAddress,
			Authorization: auth,
			Expiration:    nil,
		})
	}
	return grants
}

func addSpendingGrants(grants []authz.GrantAuthorization, info ObserverInfoReader) []authz.GrantAuthorization {
	spendMaxTokens, ok := sdkmath.NewIntFromString(info.SpendMaxTokens)
	if !ok {
		panic("Failed to parse spend max tokens")
	}
	spendAuth, err := codectypes.NewAnyWithValue(&banktypes.SendAuthorization{
		SpendLimit: sdk.NewCoins(sdk.NewCoin(config.BaseDenom, spendMaxTokens)),
	})
	if err != nil {
		panic(err)
	}
	grants = append(grants, authz.GrantAuthorization{
		Granter:       info.ObserverAddress,
		Grantee:       info.SpendGranteeAddress,
		Authorization: spendAuth,
		Expiration:    nil,
	})
	return grants
}

func addStakingGrants(grants []authz.GrantAuthorization, info ObserverInfoReader) []authz.GrantAuthorization {
	stakingMaxTokens, ok := sdkmath.NewIntFromString(info.StakingMaxTokens)
	if !ok {
		panic("Failed to parse staking max tokens")
	}
	alllowList := stakingtypes.StakeAuthorization_AllowList{AllowList: &stakingtypes.StakeAuthorization_Validators{Address: info.StakingValidatorAllowList}}

	stakingAuth, err := codectypes.NewAnyWithValue(&stakingtypes.StakeAuthorization{
		MaxTokens:         &sdk.Coin{Denom: config.BaseDenom, Amount: stakingMaxTokens},
		Validators:        &alllowList,
		AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
	})
	if err != nil {
		panic(err)
	}
	grants = append(grants, authz.GrantAuthorization{
		Granter:       info.ObserverAddress,
		Grantee:       info.StakingGranteeAddress,
		Authorization: stakingAuth,
		Expiration:    nil,
	})
	delAuth, err := codectypes.NewAnyWithValue(&stakingtypes.StakeAuthorization{
		MaxTokens:         &sdk.Coin{Denom: config.BaseDenom, Amount: stakingMaxTokens},
		Validators:        &alllowList,
		AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE,
	})
	if err != nil {
		panic(err)
	}
	grants = append(grants, authz.GrantAuthorization{
		Granter:       info.ObserverAddress,
		Grantee:       info.StakingGranteeAddress,
		Authorization: delAuth,
		Expiration:    nil,
	})
	reDelauth, err := codectypes.NewAnyWithValue(&stakingtypes.StakeAuthorization{
		MaxTokens:         &sdk.Coin{Denom: config.BaseDenom, Amount: stakingMaxTokens},
		Validators:        &alllowList,
		AuthorizationType: stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE,
	})
	if err != nil {
		panic(err)
	}
	grants = append(grants, authz.GrantAuthorization{
		Granter:       info.ObserverAddress,
		Grantee:       info.StakingGranteeAddress,
		Authorization: reDelauth,
		Expiration:    nil,
	})
	return grants

}

func AddGenesisAccount(clientCtx client.Context, balances []banktypes.Balance, appState map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	var genAccount authtypes.GenesisAccount
	totalBalanceAdded := sdk.Coins{}
	genAccounts := make([]authtypes.GenesisAccount, len(balances))
	for i, balance := range balances {
		totalBalanceAdded = totalBalanceAdded.Add(balance.Coins...)
		accAddress := sdk.MustAccAddressFromBech32(balance.Address)
		baseAccount := authtypes.NewBaseAccount(accAddress, nil, 0, 0)
		genAccount = &ethermint.EthAccount{
			BaseAccount: baseAccount,
			CodeHash:    ethcommon.BytesToHash(evmtypes.EmptyCodeHash).Hex(),
		}
		if err := genAccount.Validate(); err != nil {
			return appState, fmt.Errorf("failed to validate new genesis account: %w", err)
		}
		genAccounts[i] = genAccount
	}

	authGenState := authtypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return appState, fmt.Errorf("failed to get accounts from any: %w", err)
	}

	for _, genAc := range genAccounts {
		addr := genAc.GetAddress()
		if accs.Contains(addr) {
			return appState, fmt.Errorf("cannot add account at existing address %s", addr)
		}
		accs = append(accs, genAc)
		accs = authtypes.SanitizeGenesisAccounts(accs)
	}

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return appState, fmt.Errorf("failed to convert accounts into any's: %w", err)
	}
	authGenState.Accounts = genAccs

	authGenStateBz, err := clientCtx.Codec.MarshalJSON(&authGenState)
	if err != nil {
		return appState, fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}
	appState[authtypes.ModuleName] = authGenStateBz
	bankGenState := banktypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances...)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)
	bankGenState.Supply = bankGenState.Supply.Add(totalBalanceAdded...)

	bankGenStateBz, err := clientCtx.Codec.MarshalJSON(bankGenState)
	if err != nil {
		return appState, fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	appState[banktypes.ModuleName] = bankGenStateBz

	return appState, nil
}

func isValidatorOnly(isObserver string) bool {
	if isObserver == "y" {
		return false
	} else if isObserver == "n" {
		return true
	}
	panic("Invalid Input for isObserver field, Check observer_info.json file")
}
