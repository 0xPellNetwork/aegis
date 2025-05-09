package querytests

import (
	"fmt"

	tmcli "github.com/cometbft/cometbft/libs/cli"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xPellNetwork/aegis/x/relayer/client/cli"
	"github.com/0xPellNetwork/aegis/x/relayer/types"
	pellrelayertypes "github.com/0xPellNetwork/aegis/x/relayer/types"
)

func (s *CliTestSuite) TestShowNodeAccount() {
	ctx := s.network.Validators[0].ClientCtx
	objs := s.observerState.NodeAccountList
	common := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}
	for _, tc := range []struct {
		desc string
		id   string
		args []string
		err  error
		obj  *pellrelayertypes.NodeAccount
	}{
		{
			desc: "found",
			id:   objs[0].Operator,
			args: common,
			obj:  objs[0],
		},
		{
			desc: "not found",
			id:   "not_found",
			args: common,
			err:  status.Error(codes.InvalidArgument, "not found"),
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			args := []string{tc.id}
			args = append(args, tc.args...)
			out, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdShowNodeAccount(), args)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				s.Require().True(ok)
				s.Require().ErrorIs(stat.Err(), tc.err)
			} else {
				s.Require().NoError(err)
				var resp types.QueryNodeAccountResponse
				s.Require().NoError(s.network.Config.Codec.UnmarshalJSON(out.Bytes(), &resp))
				s.Require().NotNil(resp.NodeAccount)
				s.Require().Equal(tc.obj, resp.NodeAccount)
			}
		})
	}
}
