package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/0xPellNetwork/aegis/testutil/keeper"
	emissionstypes "github.com/0xPellNetwork/aegis/x/emissions/types"
)

func TestKeeper_GetParams(t *testing.T) {
	tests := []struct {
		name    string
		params  emissionstypes.Params
		isPanic string
	}{
		{
			name: "Successfully set params",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "",
		},
		{
			name: "negative observer slashed amount",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "slash amount cannot be less than 0",
		},
		{
			name: "MaxBondFactor too high",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.35",
				MinBondFactor:               "0.85",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "max bond factor cannot be higher that 0.25",
		},
		{
			name: "MinBondFactor too low",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.35",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "min bond factor cannot be lower that 0.75",
		},
		{
			name: "invalid block time",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "invalidTime",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "invalid block time",
		},
		{
			name: "invalid block time less than 0",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "-2",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "block time cannot be less than or equal to 0",
		},
		{
			name: "bond ratio too high",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "2.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "target bond ratio cannot be more than 100 percent",
		},
		{
			name: "bond ratio too low",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "-1.00",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "target bond ratio cannot be less than 0 percent",
		},
		{
			name: "validator emission percentage too high",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "1.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "validator emission percentage cannot be more than 100 percent",
		},
		{
			name: "validator emission percentage too low",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "-1.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "validator emission percentage cannot be less than 0 percent",
		},
		{
			name: "observer percentage too low",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "-00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "observer emission percentage cannot be less than 0 percent",
		},
		{
			name: "observer percentage too high",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "150.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "observer emission percentage cannot be more than 100 percent",
		},
		{
			name: "tss signer percentage too high",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "102.22",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "tss emission percentage cannot be more than 100 percent",
		},
		{
			name: "tss signer percentage too low",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "-102.22",
				TssGasEmissionPercentage:    "00.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "tss emission percentage cannot be less than 0 percent",
		},
		{
			name: "tss gas emission percentage too high",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "150.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "tss gas emission percentage cannot be more than 100 percent",
		},
		{
			name: "tss gas emission percentage too low",
			params: emissionstypes.Params{
				MaxBondFactor:               "1.25",
				MinBondFactor:               "0.75",
				AvgBlockTime:                "6.00",
				TargetBondRatio:             "00.67",
				ValidatorEmissionPercentage: "00.50",
				ObserverEmissionPercentage:  "00.15",
				TssSignerEmissionPercentage: "00.25",
				TssGasEmissionPercentage:    "-10.10",
				DurationFactorConstant:      "0.001877876953694702",
			},
			isPanic: "tss gas emission percentage cannot be less than 0 percent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx, _, _ := keepertest.EmissionsKeeper(t)
			assertPanic(t, func() {
				k.SetParams(ctx, tt.params)
			}, tt.isPanic)

			if tt.isPanic != "" {
				require.Equal(t, emissionstypes.DefaultParams(), k.GetParamsIfExists(ctx))
			} else {
				require.Equal(t, tt.params, k.GetParamsIfExists(ctx))
			}
		})
	}
}

func assertPanic(t *testing.T, f func(), errorLog string) {
	defer func() {
		r := recover()
		if r != nil {
			require.Contains(t, r, errorLog)
		}
	}()
	f()
}
