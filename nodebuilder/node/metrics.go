package node

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	logging "github.com/ipfs/go-log/v2"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	Meter = global.MeterProvider().Meter("node")
	log   = logging.Logger("node")
)

// WithMetrics registers node metrics.
func WithMetrics() error {
	nodeStartTS, err := Meter.
		AsyncFloat64().
		Gauge(
			"node_start_ts",
			instrument.WithDescription("timestamp when the node was started"),
		)
	if err != nil {
		log.Error(err)
		return err
	}

	totalNodeRunTime, err := Meter.
		AsyncFloat64().
		Counter(
			"node_runtime_counter_in_seconds",
			instrument.WithDescription("total time the node has been running"),
			instrument.WithUnit("seconds"),
		)
	if err != nil {
		log.Error(err)
		return err
	}

	var (
		started                  = false
		totalNodeUpTimeInSeconds = time.Now().Unix()
	)

	err = Meter.RegisterCallback(
		[]instrument.Asynchronous{nodeStartTS, totalNodeRunTime},
		func(ctx context.Context) {
			fmt.Println("I wanna see you fire up!")
			if !started {
				// Observe node start timestamp
				nodeStartTS.Observe(ctx, float64(time.Now().UTC().Unix()))
				started = true
			}

			now := time.Now().Unix()
			last := atomic.SwapInt64(&totalNodeUpTimeInSeconds, now)

			totalNodeRunTime.Observe(ctx, float64(now-last))
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
