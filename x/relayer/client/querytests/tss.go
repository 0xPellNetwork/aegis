package querytests

import (
	"fmt"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/relayer/client/cli"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (s *CliTestSuite) TestShowTSS() {
	ctx := s.network.Validators[0].ClientCtx
	obj := s.observerState.Tss
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		args []string
		err  error
		obj  *types.TSS
	}{
		{
			desc: "get",
			args: common,
			obj:  obj,
			err:  nil,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			var args []string
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowTSS(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				s.Require().True(ok)
				s.Require().ErrorIs(stat.Err(), tc.err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryTSSResponse
				s.Require().NoError(s.network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				s.Require().NotNil(resp.Tss)
				s.Require().Equal(*tc.obj, resp.Tss)
			}
		})
	}
}
