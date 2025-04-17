package txserver

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"cosmossdk.io/math"
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"
	ethermintCryptoTypes "github.com/evmos/ethermint/crypto/codec"
	"github.com/evmos/ethermint/crypto/hd"
	etherminttypes "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/pell-chain/pellcore/cmd/pellcored/config"
	"github.com/pell-chain/pellcore/e2e/utils"
	authoritytypes "github.com/pell-chain/pellcore/x/authority/types"
	emissionstypes "github.com/pell-chain/pellcore/x/emissions/types"
	lightclienttypes "github.com/pell-chain/pellcore/x/lightclient/types"
	pevmtypes "github.com/pell-chain/pellcore/x/pevm/types"
	observertypes "github.com/pell-chain/pellcore/x/relayer/types"
	restakingtypes "github.com/pell-chain/pellcore/x/restaking/types"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
	xsecuritytypes "github.com/pell-chain/pellcore/x/xsecurity/types"
)

// EmissionsPoolAddress is the address of the emissions pool
// This address is constant for all networks because it is derived from emissions name
const EmissionsPoolAddress = "pell1w43fn2ze2wyhu5hfmegr6vp52c3dgn0s9yy4j2"

// PellTxServer is a PellChain tx server for E2E test
type PellTxServer struct {
	clientCtx    client.Context
	txFactory    tx.Factory
	name         []string
	mnemonic     []string
	address      []string
	blockTimeout time.Duration
}

// NewPellTxServer returns a new TxServer with provided account
func NewTxServer(rpcAddr string, names []string, mnemonics []string, chainID string) (PellTxServer, error) {
	if len(names) == 0 {
		return PellTxServer{}, errors.New("no account provided")
	}

	if len(names) != len(mnemonics) {
		return PellTxServer{}, errors.New("invalid names and mnemonics")
	}

	// initialize rpc and check status
	rpc, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		return PellTxServer{}, fmt.Errorf("failed to initialize rpc: %s", err.Error())
	}

	if _, err = rpc.Status(context.Background()); err != nil {
		return PellTxServer{}, fmt.Errorf("failed to query rpc: %s", err.Error())
	}

	// initialize codec
	cdc, reg := newCodec()

	// initialize keyring
	kr, err := keyring.New("e2e", keyring.BackendMemory, "", os.Stdin, cdc, hd.EthSecp256k1Option())
	if err != nil {
		panic(err)
	}

	addresses := make([]string, 0, len(names))

	// create accounts
	for i := range names {
		if !bip39.IsMnemonicValid(mnemonics[i]) {
			continue
		}

		r, err := kr.NewAccount(names[i], mnemonics[i], "", "m/44'/60'/0'/0/0", hd.EthSecp256k1)
		if err != nil {
			return PellTxServer{}, fmt.Errorf("failed to create account: %s", err.Error())
		}

		accAddr, err := r.GetAddress()
		if err != nil {
			return PellTxServer{}, fmt.Errorf("failed to get account address: %s", err.Error())
		}

		addresses = append(addresses, accAddr.String())
	}

	clientCtx := newContext(rpc, cdc, reg, kr, chainID)
	txf := newFactory(clientCtx)

	return PellTxServer{
		clientCtx:    clientCtx,
		txFactory:    txf,
		name:         names,
		mnemonic:     mnemonics,
		address:      addresses,
		blockTimeout: 1 * time.Minute,
	}, nil
}

// GetAccountName returns the account name from the given index
// returns empty string if index is out of bound, error should be handled by caller
func (zts PellTxServer) GetAccountName(index int) string {
	if index >= len(zts.name) {
		return ""
	}

	return zts.name[index]
}

// GetAccountAddress returns the account address from the given index
// returns empty string if index is out of bound, error should be handled by caller
func (zts PellTxServer) GetAccountAddress(index int) string {
	if index >= len(zts.address) {
		return ""
	}
	return zts.address[index]
}

// GetAccountAddressFromName returns the account address from the given name
func (zts PellTxServer) GetAccountAddressFromName(name string) (string, error) {
	acc, err := zts.clientCtx.Keyring.Key(name)
	if err != nil {
		return "", err
	}

	addr, err := acc.GetAddress()
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

// GetAllAccountAddress returns all account addresses
func (zts PellTxServer) GetAllAccountAddress() []string {
	return zts.address
}

// GetAccountMnemonic returns the account name from the given index
// returns empty string if index is out of bound, error should be handled by caller
func (zts PellTxServer) GetAccountMnemonic(index int) string {
	if index >= len(zts.mnemonic) {
		return ""
	}

	return zts.mnemonic[index]
}

// BroadcastTx broadcasts a tx to PellChain with the provided msg from the account
// and waiting for blockTime for tx to be included in the block
func (zts PellTxServer) BroadcastTx(account string, msg sdktypes.Msg) (*sdktypes.TxResponse, error) {
	acc, err := zts.clientCtx.Keyring.Key(account)
	if err != nil {
		return nil, err
	}
	addr, err := acc.GetAddress()
	if err != nil {
		return nil, err
	}

	// set sender info
	zts.clientCtx = zts.clientCtx.
		WithFromName(account).
		WithFromAddress(addr)

	// get the newest seq and account number with retry
	var accountNumber, accountSeq uint64
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		accountNumber, accountSeq, err = zts.getLatestAccountSequence(addr)
		if err == nil {
			break
		}
		fmt.Printf("Retry %d/%d: Failed to get sequence: %v\n", i+1, maxRetries, err)
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account sequence after %d retries: %w", maxRetries, err)
	}

	zts.txFactory = zts.txFactory.
		WithAccountNumber(accountNumber).
		WithSequence(accountSeq)

	txBuilder, err := zts.txFactory.BuildUnsignedTx(msg)
	if err != nil {
		return nil, err
	}

	if err = tx.Sign(context.Background(), zts.txFactory, account, txBuilder, true); err != nil {
		return nil, err
	}

	txBytes, err := zts.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return broadcastWithBlockTimeout(zts, txBytes)
}

// broadcastWithBlockTimeout broadcasts a tx to PellChain with the provided msg from the account
// and waiting for blockTime for tx to be included in the block
// retry for sequence number errors
func broadcastWithBlockTimeout(zts PellTxServer, txBytes []byte) (*sdktypes.TxResponse, error) {
	var res *sdktypes.TxResponse
	var err error
	maxRetries := 10
	retryDelay := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		res, err = zts.clientCtx.BroadcastTx(txBytes)
		if err != nil {
			fmt.Printf("BroadcastTx failed: %v\n", err)

			if res == nil {
				time.Sleep(retryDelay)
				break
			}

			return &sdktypes.TxResponse{
				Code:      res.Code,
				Codespace: res.Codespace,
				TxHash:    res.TxHash,
			}, err
		}

		if res.Code != 0 {
			if strings.Contains(res.RawLog, "incorrect account sequence") {
				fmt.Printf("Incorrect sequence, retry %d/%d after %v...\n", i+1, maxRetries, retryDelay)
				time.Sleep(retryDelay)
				continue
			}

			return res, fmt.Errorf("broadcast failed with code %d: %s", res.Code, res.RawLog)
		}

		break
	}

	exitAfter := time.After(zts.blockTimeout)
	hash, err := hex.DecodeString(res.TxHash)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-exitAfter:
			return nil, fmt.Errorf("timed out after waiting %v for tx to be included in block", zts.blockTimeout)
		case <-time.After(time.Millisecond * 500):
			resTx, err := zts.clientCtx.Client.Tx(context.Background(), hash, false)
			if err == nil {
				return mkTxResult(zts.clientCtx, resTx)
			}
		}
	}
}

func mkTxResult(clientCtx client.Context, resTx *coretypes.ResultTx) (*sdktypes.TxResponse, error) {
	txb, err := clientCtx.TxConfig.TxDecoder()(resTx.Tx)
	if err != nil {
		return nil, err
	}
	p, ok := txb.(intoAny)
	if !ok {
		return nil, fmt.Errorf("expecting a type implementing intoAny, got: %T", txb)
	}
	resBlock, err := clientCtx.Client.Block(context.TODO(), &resTx.Height)
	if err != nil {
		return nil, err
	}
	return sdktypes.NewResponseResultTx(resTx, p.AsAny(), resBlock.Block.Time.Format(time.RFC3339)), nil
}

type intoAny interface {
	AsAny() *codectypes.Any
}

// EnableVerificationFlags enables the verification flags for the lightclient module
func (zts PellTxServer) EnableVerificationFlags(account string) error {
	// retrieve account
	acc, err := zts.clientCtx.Keyring.Key(account)
	if err != nil {
		return err
	}
	addr, err := acc.GetAddress()
	if err != nil {
		return err
	}

	_, err = zts.BroadcastTx(account, lightclienttypes.NewMsgUpdateVerificationFlags(
		addr.String(),
		true,
		true,
	))

	return err
}

// FundEmissionsPool funds the emissions pool with the given amount
func (zts PellTxServer) FundEmissionsPool(account string, amount *big.Int) error {
	// retrieve account
	acc, err := zts.clientCtx.Keyring.Key(account)
	if err != nil {
		return err
	}
	addr, err := acc.GetAddress()
	if err != nil {
		return err
	}

	// retrieve account address
	emissionPoolAccAddr, err := sdktypes.AccAddressFromBech32(EmissionsPoolAddress)
	if err != nil {
		return err
	}

	// convert amount
	amountInt := math.NewIntFromBigInt(amount)

	// fund emissions pool
	_, err = zts.BroadcastTx(account, banktypes.NewMsgSend(
		addr,
		emissionPoolAccAddr,
		sdktypes.NewCoins(sdktypes.NewCoin(config.BaseDenom, amountInt)),
	))
	return err
}

// newCodec returns the codec for msg server
func newCodec() (*codec.ProtoCodec, codectypes.InterfaceRegistry) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	sdktypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	authz.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)
	slashingtypes.RegisterInterfaces(interfaceRegistry)
	upgradetypes.RegisterInterfaces(interfaceRegistry)
	distrtypes.RegisterInterfaces(interfaceRegistry)
	evidencetypes.RegisterInterfaces(interfaceRegistry)
	crisistypes.RegisterInterfaces(interfaceRegistry)
	evmtypes.RegisterInterfaces(interfaceRegistry)
	etherminttypes.RegisterInterfaces(interfaceRegistry)
	xmsgtypes.RegisterInterfaces(interfaceRegistry)
	emissionstypes.RegisterInterfaces(interfaceRegistry)
	pevmtypes.RegisterInterfaces(interfaceRegistry)
	observertypes.RegisterInterfaces(interfaceRegistry)
	lightclienttypes.RegisterInterfaces(interfaceRegistry)
	authoritytypes.RegisterInterfaces(interfaceRegistry)
	ethermintCryptoTypes.RegisterInterfaces(interfaceRegistry)
	restakingtypes.RegisterInterfaces(interfaceRegistry)
	xsecuritytypes.RegisterInterfaces(interfaceRegistry)
	return cdc, interfaceRegistry
}

// newContext returns the client context for msg server
func newContext(
	rpc *rpchttp.HTTP,
	cdc *codec.ProtoCodec,
	reg codectypes.InterfaceRegistry,
	kr keyring.Keyring,
	chainID string,
) client.Context {
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	return client.Context{}.
		WithChainID(chainID).
		WithInterfaceRegistry(reg).
		WithCodec(cdc).
		WithTxConfig(txConfig).
		WithLegacyAmino(codec.NewLegacyAmino()).
		WithInput(os.Stdin).
		WithOutput(os.Stdout).
		WithBroadcastMode(flags.BroadcastSync).
		WithClient(rpc).
		WithSkipConfirmation(true).
		WithKeyring(kr).
		WithAccountRetriever(authtypes.AccountRetriever{})
}

// newFactory returns the tx factory for msg server
func newFactory(clientCtx client.Context) tx.Factory {
	return tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(40000000).
		WithGasAdjustment(1).
		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithTxConfig(clientCtx.TxConfig).
		WithFees("400000000000000000apell")
}

func GetEventAttribute(resp *sdktypes.TxResponse, eventType, attributeKey string) (string, error) {
	for _, event := range resp.Events {
		if event.Type == eventType {
			for _, attri := range event.Attributes {
				if attri.Key == attributeKey {
					addr := attri.Value
					addr = addr[1 : len(addr)-1]
					return addr, nil
				}
			}

			return "", fmt.Errorf("attribute %s not found, attributes:  %+v", eventType, attributeKey)
		}
	}

	return "", fmt.Errorf("attribute %s not found, attributes:  %+v", eventType, attributeKey)
}

func (ts *PellTxServer) DeployPellSystemContract(admin string) (*sdktypes.TxResponse, error) {
	acc, err := ts.clientCtx.Keyring.Key(admin)
	if err != nil {
		return nil, err
	}

	adminAddr, err := acc.GetAddress()
	if err != nil {
		return nil, err
	}

	return ts.BroadcastTx(admin, pevmtypes.NewMsgDeploySystemContracts(adminAddr.String()))
}

func (zts PellTxServer) SendPellFromAdmin(to sdktypes.AccAddress, amount *big.Int) error {
	adminAcc, err := zts.clientCtx.Keyring.Key(utils.FungibleAdminName)
	if err != nil {
		return err
	}
	adminAddr, err := adminAcc.GetAddress()
	if err != nil {
		return err
	}

	amountInt := math.NewIntFromBigInt(amount.Mul(amount, big.NewInt(1e18)))

	resp, err := zts.BroadcastTx(utils.FungibleAdminName, banktypes.NewMsgSend(
		adminAddr,
		to,
		sdktypes.NewCoins(sdktypes.NewCoin(config.BaseDenom, amountInt)),
	))
	if err != nil {
		panic(err)
	}

	if resp.Code != 0 {
		panic(resp)
	}

	return nil
}

func (zts PellTxServer) getLatestAccountSequence(addr sdktypes.AccAddress) (uint64, uint64, error) {
	// get account seq and account number from chain
	queryClient := authtypes.NewQueryClient(zts.clientCtx)
	res, err := queryClient.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: addr.String(),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query account: %w", err)
	}

	var acc authtypes.AccountI
	err = zts.clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to unpack account: %w", err)
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
