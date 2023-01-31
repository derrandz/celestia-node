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
func WithMetrics() {
	nodeStartTS, err := meter.
		AsyncFloat64().
		Gauge(
			"node_start_ts",
			instrument.WithDescription("timestamp when the node was started"),
		)
	if err != nil {
		log.Fatal(err)
	}

	totalNodeRunTime, err := meter.
		AsyncFloat64().
		Counter(
			"node_runtime_counter_in_seconds",
			instrument.WithDescription("total time the node has been running"),
			instrument.WithUnit("seconds"),
		)
	if err != nil {
		log.Fatal(err)
	}

	var started bool = false
	var totalNodeUpTimeInSeconds int64 // Observe total node run time
	err = meter.RegisterCallback(
		[]instrument.Asynchronous{totalNodeRunTime},
		func(ctx context.Context) {
			if !started {
				// Observe node start timestamp
				nodeStartTS.Observe(context.Background(), float64(time.Now().UTC().Unix()))
				started = true
			}

			now := time.Now().Unix()
			last := atomic.SwapInt64(&totalNodeUpTimeInSeconds, now)

			totalNodeRunTime.Observe(ctx, float64(now-last))
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
