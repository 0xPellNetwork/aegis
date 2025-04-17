package common

import (
	"context"
	"time"

	"github.com/pell-chain/pellcore/e2e/runner"
	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

const WAIT_KEYGEN_HEIGHT = 15

// waitKeygenHeight waits for keygen height
func WaitKeygenHeight(
	ctx context.Context,
	xmsgClient xmsgtypes.QueryClient,
	logger *runner.Logger,
) {
	logger.Print("⏳ wait height %v for keygen to be completed", WAIT_KEYGEN_HEIGHT)

	for {
		response, err := xmsgClient.LastPellHeight(ctx, &xmsgtypes.QueryLastPellHeightRequest{})
		if err != nil {
			logger.Error("xmsgClient.LastPellHeight error: %s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if response.Height >= WAIT_KEYGEN_HEIGHT {
			break
		}

		time.Sleep(2 * time.Second)
		logger.Info("Last PellHeight: %d", response.Height)
	}
}
