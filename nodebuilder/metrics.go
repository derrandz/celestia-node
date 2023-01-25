// This file defines metrics relative to the nodebuilder package.
package nodebuilder

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.uber.org/fx"
)

type metrics struct {
	// nodeStartTS is the timestamp when the node was started.
	nodeStartTS syncfloat64.UpDownCounter

	// totalNodeUptime is the total time the node has been running.
	totalNodeUptime syncfloat64.Counter

	// lastNodeUptimeTs is the last timestamp when the node uptime was recorded.
	lastNodeUptimeTs float64
}

var (
	meter = global.MeterProvider().Meter("node")
)

func newNodeMetrics() (*metrics, error) {
	nodeStartTS, err := meter.
		SyncFloat64().
		UpDownCounter(
			"node_start_ts",
			instrument.WithDescription("timestamp when the node was started"),
		)
	if err != nil {
		return nil, err
	}

	totalNodeUptime, err := meter.
		SyncFloat64().
		Counter(
			"node_uptime",
			instrument.WithDescription("total time the node has been running"),
		)
	if err != nil {
		return nil, err
	}

	return &metrics{
		nodeStartTS:     nodeStartTS,
		totalNodeUptime: totalNodeUptime,
	}, nil
}

// recordNodeStart records the timestamp when the node was started.
func (m *metrics) recordNodeStart(ctx context.Context) {
	m.nodeStartTS.Add(context.Background(), float64(time.Now().Unix()))
}

// recordNodeUptime records the total time the node has been running.
func (m *metrics) recordNodeUptime(ctx context.Context, interval time.Duration) {
	m.lastNodeUptimeTs = float64(time.Now().Unix())

	// ticker ticks every `interval` and records the total time the node has been running
	// since the last tick
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			ts := time.Since(time.Unix(int64(m.lastNodeUptimeTs), 0)).Seconds()
			m.lastNodeUptimeTs = ts
			m.totalNodeUptime.Add(context.Background(), ts)
		case <-ctx.Done():
			return
		}
	}
}

// WithNodeMetrics returns a function that initializes the node metrics
// and registers onStart fx hook to record the node uptime and register
// a callback that records the total node uptime every other `NodeUptimeScrapeInterval`
// see `nodebuilder/telemetry/config.go`.
func WithNodeMetrics(lifecycleFunc fx.Lifecycle, node *Node, cfg *Config) {
	m, err := newNodeMetrics()
	if err != nil {
		panic(err)
	}

	node.metrics = m

	lifecycleFunc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				node.metrics.recordNodeStart(context.Background())

				interval := time.Duration(cfg.Telemetry.NodeUptimeScrapeInterval) * time.Minute
				go node.metrics.recordNodeUptime(context.Background(), interval)

				return nil
			},
		},
	)
}
