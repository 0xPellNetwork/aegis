package e2etests

import (
	"crypto/rand"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/0xPellNetwork/contracts/pkg/contracts/pell_evm/registry/registryrouter.sol"
	"github.com/0xPellNetwork/pell-middleware-contracts/pkg/src/mocks/mockdvsservicemanager.sol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/0xPellNetwork/aegis/e2e/runner"
	"github.com/0xPellNetwork/aegis/pkg/crypto/bls"
	xsecuritytypes "github.com/0xPellNetwork/aegis/x/xsecurity/types"
)

// generateBLSPubkeyParams returns the BLS public key parameters for the test case
func generateBLSPubkeyParams(operatorAddr common.Address, chain *runner.EVMChain) registryrouter.IRegistryRouterPubkeyRegistrationParams {
	g1HashedMsgToSign, err := chain.EvmContracts.RegistryRouter.PubkeyRegistrationMessageHash(&bind.CallOpts{}, operatorAddr)
	if err != nil {
		panic(err)
	}

	return generateBLSPubkeyParamsBySign(g1HashedMsgToSign)
}

// generateBLSPubkeyParamsBySign returns the BLS public key parameters for the test case
func generateBLSPubkeyParamsBySign(g1HashedMsgToSign registryrouter.BN254G1Point) registryrouter.IRegistryRouterPubkeyRegistrationParams {
	blsKeyPair, err := bls.GenRandomBlsKeys()
	if err != nil {
		panic(err)
	}

	g1hashed := bls.NewG1Point(g1HashedMsgToSign.X, g1HashedMsgToSign.Y).G1Affine
	signedMsg := convertToBN254G1Point(
		blsKeyPair.SignHashedToCurveMessage(g1hashed).G1Point,
	)

	pubkeyRegistrationSignature := registryrouter.BN254G1Point{
		X: signedMsg.X,
		Y: signedMsg.Y,
	}

	G1pubkeyBN254 := convertToBN254G1Point(blsKeyPair.GetPubKeyG1())
	G2pubkeyBN254 := convertToBN254G2Point(blsKeyPair.GetPubKeyG2())

	pg1 := registryrouter.BN254G1Point{
		X: G1pubkeyBN254.X,
		Y: G1pubkeyBN254.Y,
	}
	pg2 := registryrouter.BN254G2Point{
		X: G2pubkeyBN254.X,
		Y: G2pubkeyBN254.Y,
	}

	pubkeyRegParams := registryrouter.IRegistryRouterPubkeyRegistrationParams{
		PubkeyRegistrationSignature: pubkeyRegistrationSignature,
		PubkeyG1:                    pg1,
		PubkeyG2:                    pg2,
	}
	return pubkeyRegParams
}

func generateRandomSalt() [32]byte {
	var salt [32]byte
	if _, err := rand.Read(salt[:]); err != nil {
		panic(err)
	}

	return salt
}

func convertToBN254G1Point(input *bls.G1Point) mockdvsservicemanager.BN254G1Point {
	output := mockdvsservicemanager.BN254G1Point{
		X: input.X.BigInt(big.NewInt(0)),
		Y: input.Y.BigInt(big.NewInt(0)),
	}
	return output
}

func convertToBN254G2Point(input *bls.G2Point) mockdvsservicemanager.BN254G2Point {
	output := mockdvsservicemanager.BN254G2Point{
		X: [2]*big.Int{input.X.A1.BigInt(big.NewInt(0)), input.X.A0.BigInt(big.NewInt(0))},
		Y: [2]*big.Int{input.Y.A1.BigInt(big.NewInt(0)), input.Y.A0.BigInt(big.NewInt(0))},
	}
	return output
}

// ConvertPubkeyRegistrationParamsFromEventToStore LST helper function to convert pubkey registration params from event to store
func ConvertPubkeyRegistrationParamsFromEventToStore(params registryrouter.IRegistryRouterPubkeyRegistrationParams) *xsecuritytypes.PubkeyRegistrationParams {
	return &xsecuritytypes.PubkeyRegistrationParams{
		PubkeyRegistrationSignature: &xsecuritytypes.G1Point{
			X: sdkmath.NewIntFromBigInt(params.PubkeyRegistrationSignature.X),
			Y: sdkmath.NewIntFromBigInt(params.PubkeyRegistrationSignature.Y),
		},
		PubkeyG1: &xsecuritytypes.G1Point{
			X: sdkmath.NewIntFromBigInt(params.PubkeyG1.X),
			Y: sdkmath.NewIntFromBigInt(params.PubkeyG1.Y),
		},
		PubkeyG2: &xsecuritytypes.G2Point{
			X: []sdkmath.Int{
				sdkmath.NewIntFromBigInt(params.PubkeyG2.X[0]),
				sdkmath.NewIntFromBigInt(params.PubkeyG2.X[1]),
			},
			Y: []sdkmath.Int{
				sdkmath.NewIntFromBigInt(params.PubkeyG2.Y[0]),
				sdkmath.NewIntFromBigInt(params.PubkeyG2.Y[1]),
			},
		},
	}
}
