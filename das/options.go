package das

import (
	"fmt"
	"time"
)

var (
	ErrInvalidOption      = fmt.Errorf("das: invalid option")
	ErrInvalidOptionValue = func(optionName string, value string) error {
		return fmt.Errorf("%w: value %s cannot be %s", ErrInvalidOption, optionName, value)
	}
)

// Type Option is the functional option that is applied to the daser instance for parameters configuration.
type Option func(*DASer) error

// Type parameters is the set of parameters that must be configured for the daser
type parameters struct {
	samplingRange     uint
	concurrencyLimit  uint
	bgStoreInterval   time.Duration
	priorityQueueSize uint
	genesisHeight     uint
}

// defaultParameters returns the default configuration values for the daser parameters
func defaultParameters() parameters {
	return parameters{
		samplingRange:     100,
		concurrencyLimit:  16,
		bgStoreInterval:   10 * time.Minute,
		priorityQueueSize: 16 * 4,
		genesisHeight:     1,
	}
}

// WithSamplingRange is a functional option to configure the daser's `samplingRange` parameter
// ```
//
//	WithSamplingRange(10)(daser)
//
// ```
// or
// ```
//
//	option := WithSamplingRange(10)
//	// shenanigans to create daser
//	option(daser)
//
// ```
func WithSamplingRange(samplingRange uint) Option {
	return func(d *DASer) error {
		if samplingRange == 0 {
			return ErrInvalidOptionValue(
				"samplingRange",
				"0",
			)
		}

		d.params.samplingRange = samplingRange

		return nil
	}
}

// WithConcurrencyLimit is a functional option to configure the daser's `concurrencyLimit` parameter
// Refer to WithSamplingRange documentation to see an example of how to use this
func WithConcurrencyLimit(concurrencyLimit uint) Option {
	return func(d *DASer) error {
		if concurrencyLimit == 0 {
			return ErrInvalidOptionValue(
				"concurrencyLimit",
				"0",
			)
		}
		d.params.concurrencyLimit = concurrencyLimit
		return nil
	}
}

// WithBackgroundStoreInterval is a functional option to configure the daser's `backgroundStoreINterval` parameter
// Refer to WithSamplingRange documentation to see an example of how to use this
func WithBackgroundStoreInterval(bgStoreInterval time.Duration) Option {
	return func(d *DASer) error {
		d.params.bgStoreInterval = bgStoreInterval
		return nil
	}
}

// WithPriorityQueueSize is a functional option to configure the daser's `priorityQueuSize` parameter
// Refer to WithSamplingRange documentation to see an example of how to use this
func WithPriorityQueueSize(priorityQueueSize uint) Option {
	return func(d *DASer) error {
		d.params.priorityQueueSize = priorityQueueSize
		return nil
	}
}

// WithGenesisHeight is a functional option to configure the daser's `GenesisHeight` parameter
// Refer to WithSamplingRange documentation to see an example of how to use this
func WithGenesisHeight(genesisHeight uint) Option {
	return func(d *DASer) error {
		if genesisHeight == 0 {
			return ErrInvalidOptionValue(
				"genesisHeight",
				"0",
			)
		}
		d.params.genesisHeight = genesisHeight
		return nil
	}
}
