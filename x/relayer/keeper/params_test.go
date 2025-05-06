package keeper_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func TestGetParams(t *testing.T) {
	k, ctx, _, _ := keepertest.RelayerKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParamsIfExists(ctx))
}

func TestGenerateAddress(t *testing.T) {
	addr := sdk.AccAddress(crypto.AddressHash([]byte("Output1" + strconv.Itoa(1))))
	addrString := addr.String()
	addbech32, _ := sdk.AccAddressFromBech32(addrString)
	valAddress := sdk.ValAddress(addbech32)
	v, _ := sdk.ValAddressFromBech32(valAddress.String())
	accAddress := sdk.AccAddress(v)
	a, _ := sdk.AccAddressFromBech32(accAddress.String())
	fmt.Println(a.String())
}
