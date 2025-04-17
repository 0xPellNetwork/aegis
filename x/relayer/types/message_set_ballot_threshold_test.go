package types

// FIXME: make it work
//
//func TestMsgSetBallotThreshold_ValidateBasic(t *testing.T) {
//	tests := []struct {
//		name string
//		msg  MsgSetBallotThreshold
//		err  error
//	}{
//		{
//			name: "invalid address",
//			msg: MsgSetBallotThreshold{
//				Signer: "invalid_address",
//			},
//			err: sdkerrors.ErrInvalidAddress,
//		}, {
//			name: "valid address",
//			msg: MsgSetBallotThreshold{
//				Signer: sample.AccAddress(),
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			err := tt.msg.ValidateBasic()
//			if tt.err != nil {
//				require.ErrorIs(t, err, tt.err)
//				return
//			}
//			require.NoError(t, err)
//		})
//	}
//}
