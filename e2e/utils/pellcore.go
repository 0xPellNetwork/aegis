package utils

import (
	"context"
	"fmt"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	xmsgtypes "github.com/pell-chain/pellcore/x/xmsg/types"
)

const (
	FungibleAdminName = "fungibleadmin"
	Pellcore0Name     = "pellcore0"
	Pellcore0Mnemonic = "pitch omit flag fuel soap artefact sleep hurdle segment hurry then wear plunge talk tragic huge spider open charge father filter behave ski coffee"

	DefaultXmsgTimeout = 4 * time.Minute
)

// WaitXmsgMinedByInTxHash waits until xmsg is mined; returns the xmsgIndex (the last one)
func WaitXmsgMinedByInTxHash(
	ctx context.Context,
	inTxHash string,
	xmsgClient xmsgtypes.QueryClient,
	logger infoLogger,
	xmsgTimeout time.Duration,
) *xmsgtypes.Xmsg {
	xmsgs := WaitXmsgsMinedByInTxHash(ctx, inTxHash, xmsgClient, 1, logger, xmsgTimeout)
	if len(xmsgs) == 0 {
		panic(fmt.Sprintf("xmsg not found, inTxHash: %s", inTxHash))
	}
	return xmsgs[len(xmsgs)-1]
}

// WaitXmsgsMinedByInTxHash waits until xmsg is mined; returns the xmsgIndex (the last one)
func WaitXmsgsMinedByInTxHash(
	ctx context.Context,
	inTxHash string,
	xmsgClient xmsgtypes.QueryClient,
	xmsgsCount int,
	logger infoLogger,
	xmsgTimeout time.Duration,
) []*xmsgtypes.Xmsg {
	startTime := time.Now()

	timeout := DefaultXmsgTimeout
	if xmsgTimeout != 0 {
		timeout = xmsgTimeout
	}

	// fetch xmsgs by inTxHash
	for i := 0; ; i++ {
		// declare xmsgs here so we can print the last fetched one if we reach timeout
		var xmsgs []*xmsgtypes.Xmsg

		if time.Since(startTime) > timeout {
			xmsgMessage := ""
			if len(xmsgs) > 0 {
				xmsgMessage = fmt.Sprintf(", last xmsg: %v", xmsgs[0].String())
			}

			panic(fmt.Sprintf("waiting xmsg timeout, xmsg not mined, inTxHash: %s%s", inTxHash, xmsgMessage))
		}
		time.Sleep(1 * time.Second)

		res, err := xmsgClient.InTxHashToXmsgData(ctx, &xmsgtypes.QueryInTxHashToXmsgDataRequest{
			InTxHash: inTxHash,
		})

		if err != nil {
			// prevent spamming logs
			if i%10 == 0 {
				logger.Info("Error getting xmsg by inTxHash: %s", err.Error())
			}
			continue
		}
		if len(res.Xmsgs) < xmsgsCount {
			// prevent spamming logs
			if i%10 == 0 {
				logger.Info(
					"not enough xmsgs found by inTxHash: %s, expected: %d, found: %d",
					inTxHash,
					xmsgsCount,
					len(res.Xmsgs),
				)
			}
			continue
		}
		xmsgs = make([]*xmsgtypes.Xmsg, 0, len(res.Xmsgs))
		allFound := true
		for j, xmsg := range res.Xmsgs {
			xmsg := xmsg
			if !IsTerminalStatus(xmsg.XmsgStatus.Status) {
				// prevent spamming logs
				if i%10 == 0 {
					logger.Info(
						"waiting for xmsg index %d to be mined by inTxHash: %s, xmsg status: %s, message: %s",
						j,
						inTxHash,
						xmsg.XmsgStatus.Status.String(),
						xmsg.XmsgStatus.StatusMessage,
					)
				}
				allFound = false
				break
			}
			xmsgs = append(xmsgs, &xmsg)
		}
		if !allFound {
			continue
		}
		return xmsgs
	}
}

// WaitXmsgMinedByIndex waits until xmsg is mined; returns the xmsgIndex
func WaitXmsgMinedByIndex(
	ctx context.Context,
	xmsgIndex string,
	xmsgClient xmsgtypes.QueryClient,
	logger infoLogger,
	xmsgTimeout time.Duration,
) *xmsgtypes.Xmsg {
	startTime := time.Now()

	timeout := DefaultXmsgTimeout
	if xmsgTimeout != 0 {
		timeout = xmsgTimeout
	}

	for i := 0; ; i++ {
		if time.Since(startTime) > timeout {
			panic(fmt.Sprintf(
				"waiting xmsg timeout, xmsg not mined, xmsg: %s",
				xmsgIndex,
			))
		}
		time.Sleep(1 * time.Second)

		// fetch xmsg by index
		res, err := xmsgClient.Xmsg(ctx, &xmsgtypes.QueryGetXmsgRequest{
			Index: xmsgIndex,
		})
		if err != nil {
			// prevent spamming logs
			if i%10 == 0 {
				logger.Info("Error getting xmsg by inTxHash: %s", err.Error())
			}
			continue
		}
		xmsg := res.Xmsg
		if !IsTerminalStatus(xmsg.XmsgStatus.Status) {
			// prevent spamming logs
			if i%10 == 0 {
				logger.Info(
					"waiting for xmsg to be mined from index: %s, xmsg status: %s, message: %s",
					xmsgIndex,
					xmsg.XmsgStatus.Status.String(),
					xmsg.XmsgStatus.StatusMessage,
				)
			}
			continue
		}

		return xmsg
	}
}

func IsTerminalStatus(status xmsgtypes.XmsgStatus) bool {
	return status == xmsgtypes.XmsgStatus_OUTBOUND_MINED ||
		status == xmsgtypes.XmsgStatus_ABORTED ||
		status == xmsgtypes.XmsgStatus_REVERTED
}

// WaitForBlockHeight waits until the block height reaches the given height
func WaitForBlockHeight(
	ctx context.Context,
	height int64,
	rpcURL string,
	logger infoLogger,
) {
	// initialize rpc and check status
	rpc, err := rpchttp.New(rpcURL, "/websocket")
	if err != nil {
		panic(err)
	}
	status := &coretypes.ResultStatus{}
	for i := 0; status.SyncInfo.LatestBlockHeight < height; i++ {
		status, err = rpc.Status(ctx)
		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)

		// prevent spamming logs
		if i%10 == 0 {
			logger.Info("waiting for block: %d, current height: %d\n", height, status.SyncInfo.LatestBlockHeight)
		}
	}
}
