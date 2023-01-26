package node

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/require"
)

// Test UptimeMetrics.Persist and UptimeMetrics.Get
func TestUptimeMetrics_Persist(t *testing.T) {
	ctx := context.Background()
	ds := datastore.NewMapDatastore()
	m, err := NewUptimeMetrics(ds)
	require.NoError(t, err)

	floatValue := float64(12312312.34)
	err = m.Persist(ctx, nodeStartTSKey, floatValue)
	require.NoError(t, err)

	val, err := m.Get(ctx, nodeStartTSKey)
	require.NoError(t, err)
	require.Equal(t, floatValue, val)
}
