package node

import (
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
)

// UptimeMetrics is a struct that records
//
//  1. node start time: the timestamp when the node was started
//  2. node uptime: total time the node has been running counted in units of 1 second
//
// the node start time is recorded imperatively when RecordNodeStartTime is called
// whereas the node uptime is recorded periodically
// upon callback recalling (re-mettering from optl)
type UptimeMetrics struct {
	// nodeStartTS is the timestamp when the node was started.
	nodeStartTS asyncfloat64.Gauge

	// totalNodeRunTime is the total time the node has been running.
	totalNodeRunTime asyncfloat64.Counter

	// totalNodeUpTimeTicks is the total time the node has been running in seconds
	totalNodeUpTimeInSeconds int64
}

var (
	meter = global.MeterProvider().Meter("node")
)

// NewUptimeMetrics creates a new UptimeMetrics
// and registers a callback to re-meter the totalNodeRunTime metric.
func NewUptimeMetrics() (*UptimeMetrics, error) {
	nodeStartTS, err := meter.
		AsyncFloat64().
		Gauge(
			"node_start_ts",
			instrument.WithDescription("timestamp when the node was started"),
		)
	if err != nil {
		return nil, err
	}

	totalNodeRunTime, err := meter.
		AsyncFloat64().
		Counter(
			"node_runtime_counter_in_seconds",
			instrument.WithDescription("total time the node has been running"),
			instrument.WithUnit("seconds"),
		)
	if err != nil {
		return nil, err
	}

	m := &UptimeMetrics{
		nodeStartTS:      nodeStartTS,
		totalNodeRunTime: totalNodeRunTime,
	}

	err = meter.RegisterCallback(
		[]instrument.Asynchronous{
			totalNodeRunTime,
		},
		func(ctx context.Context) {
			now := time.Now().UTC().Unix()
			old := atomic.SwapInt64(&m.totalNodeUpTimeInSeconds, now)

			totalNodeRunTime.Observe(
				ctx,
				time.Since(time.Unix(old, 0).UTC()).Seconds(),
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// RecordNodeStartTime records the timestamp when the node was started.
func (m *UptimeMetrics) RecordNodeStartTime(ctx context.Context) {
	nodeStartTS := float64(time.Now().Unix())
	m.nodeStartTS.Observe(context.Background(), nodeStartTS)
}
