package node

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func TestWithMetrics(t *testing.T) {
	exp, err := stdoutmetric.New()
	if err != nil {
		panic(err)
	}

	reader := metric.NewPeriodicReader(exp, metric.WithTimeout(1*time.Second))

	provider := metric.NewMeterProvider(
		metric.WithReader(reader),
	)

	Meter = provider.Meter("test")

	err = WithMetrics()
	require.NoError(t, err)

	mr, err := reader.Collect(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 1, len(mr.ScopeMetrics))
	assert.Equal(t, 2, len(mr.ScopeMetrics[0].Metrics))

	nodeStartTS := mr.ScopeMetrics[0].Metrics[0]
	totalNodeRunTime := mr.ScopeMetrics[0].Metrics[1]

	assert.Equal(t, nodeStartTS.Name, "node_start_ts")
	assert.Equal(t, totalNodeRunTime.Name, "node_runtime_counter_in_seconds")

	// assert.Equal(t, nodeStartTS.Data, ??)
	// assert.Equal(t, totalNodeRunTime.Data, ??)
}
