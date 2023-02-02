package node

import (
	"context"
	"sync/atomic"
	"time"

	logging "github.com/ipfs/go-log/v2"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	meter = global.MeterProvider().Meter("node")
	log   = logging.Logger("node")
)

// WithMetrics registers node metrics.
func WithMetrics() error {
	nodeStartTS, err := meter.
		AsyncFloat64().
		Gauge(
			"node_start_ts",
			instrument.WithDescription("timestamp when the node was started"),
		)
	if err != nil {
		log.Error(err)
		return err
	}

	totalNodeRunTime, err := meter.
		AsyncFloat64().
		Counter(
			"node_runtime_counter_in_seconds",
			instrument.WithDescription("total time the node has been running"),
		)
	if err != nil {
		log.Error(err)
		return err
	}

	var (
		started                  = false
		totalNodeUpTimeInSeconds = time.Now().Unix()
	)

	err = meter.RegisterCallback(
		[]instrument.Asynchronous{nodeStartTS, totalNodeRunTime},
		func(ctx context.Context) {
			if !started {
				// Observe node start timestamp
				nodeStartTS.Observe(ctx, float64(time.Now().UTC().Unix()))
				started = true
			}

			now := time.Now().Unix()
			last := atomic.SwapInt64(&totalNodeUpTimeInSeconds, now)

			totalNodeRunTime.Observe(ctx, time.Duration(now-last).Seconds())
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
