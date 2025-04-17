package pellcore

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	pctx "github.com/pell-chain/pellcore/relayer/context"
)

var logSampler = &zerolog.BasicSampler{N: 10}

// CoreContextUpdater is a polling goroutine that checks and updates core context at every height
func (b *PellCoreBridge) UpdateAppContextWorker(ctx context.Context) {
	app, err := pctx.FromContext(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			b.logger.Error().Interface("panic", r).Msg("UpdateAppContextWorker: recovered from panic")
		}
	}()

	var (
		updateEvery = time.Duration(app.Config().ConfigUpdateTicker) * time.Second
		ticker      = time.NewTicker(updateEvery)
		logger      = b.logger.Sample(logSampler)
	)

	b.logger.Info().Msg("UpdateAppContextWorker started")

	for {
		select {
		case <-ticker.C:
			b.logger.Debug().Msg("UpdateAppContextWorker invocation")
			if err := b.UpdateAppContext(ctx, app, false, logger); err != nil {
				b.logger.Err(err).Msg("UpdateAppContextWorker failed to update config")
			}
		case <-b.stop:
			b.logger.Info().Msg("UpdateAppContextWorker stopped")
			return
		}
	}
}
