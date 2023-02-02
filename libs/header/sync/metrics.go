package sync

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	meter = global.MeterProvider().Meter("header/sync")
)

type metrics struct {
	totalSynced int64
}

func (s *Syncer[H]) InitMetrics() error {
	log.Debug("Initializing syncer metrics")
	totalSynced, err := meter.
		AsyncFloat64().
		Gauge(
			"total_synced_headers",
			instrument.WithDescription("total synced headers"),
		)
	if err != nil {
		return err
	}

	err = meter.RegisterCallback(
		[]instrument.Asynchronous{
			totalSynced,
		},
		func(ctx context.Context) {
			totalSynced.Observe(ctx, float64(atomic.LoadInt64(&s.metrics.totalSynced)))
		},
	)
	if err != nil {
		return err
	}

	s.metrics = &metrics{}

	return nil
}

// recordTotalSampled records the total sampled headers.
func (m *metrics) recordTotalSynced(totalSampled int) {
	if m == nil {
		return
	}
	atomic.AddInt64(&m.totalSynced, 1)
}
