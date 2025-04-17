package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// PELLRevertAndCallContract deposits native PELL to the sender address if its account or if the account does not exist yet
// If it's not an account it calls onRevert function of the connector contract and provides the sender address as the pellTxSenderAddress.The amount of tokens is minted to the pevm module account, wrapped and sent to the contract
func (k Keeper) PELLRevertAndCallContract(ctx sdk.Context,
	sender ethcommon.Address,
	to ethcommon.Address,
	inboundSenderChainID int64,
	destinationChainID int64,
	indexBytes [32]byte) (*evmtypes.MsgEthereumTxResponse, error) {
	acc := k.evmKeeper.GetAccount(ctx, sender)
	if acc == nil || !acc.IsContract() {
		return nil, nil
	}
	// Call onRevert function of the connector contract. The connector contract will then call the onRevert function of the pellTxSender contract which is the sender address
	return nil, nil
}
