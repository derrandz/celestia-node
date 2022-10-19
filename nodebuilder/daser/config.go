package daser

import (
	"time"
)

type Config struct {
	//  samplingRange is the maximum amount of headers processed in one job.
	SamplingRange uint

	// concurrencyLimit defines the maximum amount of sampling workers running in parallel.
	ConcurrencyLimit uint

	// backgroundStoreInterval is the period of time for background checkpointStore to perform a checkpoint backup.
	BackgroundStoreInterval time.Duration

	// priorityQueueSize defines the size limit of the priority queue
	PriorityQueueSize uint

	// genesisHeight is the height sampling will start from
	GenesisHeight uint
}

// TODO(@derrandz): parameters needs performance testing on real network to define optimal values
func DefaultConfig() Config {
	return Config{
		SamplingRange:           100,
		ConcurrencyLimit:        16,
		BackgroundStoreInterval: 10 * time.Minute,
		PriorityQueueSize:       16 * 4,
		GenesisHeight:           1,
	}
}

// Validate performs basic validation of the config.
func (cfg *Config) Validate() error {
	// TODO(team): what should validate for in here? seems like all configuration fields can accept zero-values.
	return nil
}
