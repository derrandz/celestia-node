package nodebuilder

import (
	"testing"

	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	"github.com/celestiaorg/celestia-node/nodebuilder/node/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestWithMetrics(t *testing.T) {
	// instantiate gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// instantiate metrics mock
	mockMeter := mocks.NewMockMeter(ctrl)

	// instantiate InstrumentProvider from asyncfloat64.go
	instrumentProvider := mocks.NewMockInstrumentProvider(ctrl)

	// instantiate counter mock from asyncfloat64.go
	// counterMock := mocks.NewMockCounter(ctrl)

	// instantiate gauge mock from asyncfloat64.go
	// gaugeMock := mocks.NewMockGauge(ctrl)

	// define expectation on instrumentProvider.Counter() to return counterMock
	instrumentProvider.
		EXPECT().
		Counter(
			"node_runtime_counter_in_seconds",
			gomock.Any(),
			gomock.Any(),
		)
		// Return(counterMock, nil)

	// define expectation on instrumentProvider.Gauge() to return gaugeMock
	instrumentProvider.
		EXPECT().
		Gauge(
			"node_start_ts",
			gomock.Any(),
		)
		// Return(gaugeMock, nil)

	// define expectation on metricsMock.AsyncFloat64() to return instrumentProvider
	mockMeter.
		EXPECT().
		AsyncFloat64().
		Return(instrumentProvider).
		Times(2)

	// TODO(@derrandz): uncomment when weird "asynchronous method is missing" error is fixed
	// counterObservationCall := counterMock.
	// 	EXPECT().
	// 	Observe(gomock.Any(), gomock.Any()).
	// 	Times(1)
	//
	// gaugeMock.
	// 	EXPECT().
	// 	Observe(gomock.Any(), gomock.Any()).
	// 	After(counterObservationCall)

	// define expectation on metricsMock.RegisterCallback() to return nil
	mockMeter.
		EXPECT().
		RegisterCallback(gomock.Any(), gomock.Any()).
		Return(nil)

	// replace the global variable with the mock
	node.Meter = mockMeter
	err := node.WithMetrics()
	require.NoError(t, err)
}
