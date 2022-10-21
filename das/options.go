package das

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidOptionValue = func(optionName string, value string) error {
		errorMsg := fmt.Sprintf("das/daser: invalid option value: %s, cannot be %s", optionName, value)
		return errors.New(errorMsg)
	}
)

type Option func(*DASer) error

type parameters struct {
	samplingRange     uint
	concurrencyLimit  uint
	bgStoreInterval   time.Duration
	priorityQueueSize uint
	genesisHeight     uint
}

func defaultParameters() parameters {
	return parameters{
		samplingRange:     100,
		concurrencyLimit:  16,
		bgStoreInterval:   10 * time.Minute,
		priorityQueueSize: 16 * 4,
		genesisHeight:     1,
	}
}

// Option for: samplingRange
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

// Option for: concurrencyLimit
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

// Option for: bgStoreInterval
func WithBackgroundStoreInterval(bgStoreInterval time.Duration) Option {
	return func(d *DASer) error {
		d.params.bgStoreInterval = bgStoreInterval
		return nil
	}
}

// Option for: priorityQueueSize
func WithPriorityQueueSize(priorityQueueSize uint) Option {
	return func(d *DASer) error {
		d.params.priorityQueueSize = priorityQueueSize
		return nil
	}
}

// Option for: genesisHeight
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
