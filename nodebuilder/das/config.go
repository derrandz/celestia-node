package das

import (
	"errors"
	"time"
)

// Config contains configuration parameters for the DASer (or DASing process)
type Config struct {
	//  SamplingRange is the maximum amount of headers processed in one job.
	SamplingRange uint

	// ConcurrencyLimit defines the maximum amount of sampling workers running in parallel.
	ConcurrencyLimit int

	// BackgroundStoreInterval is the period of time for background checkpointStore to perform a checkpoint backup.
	BackgroundStoreInterval time.Duration

	// PriorityQueueSize defines the size limit of the priority queue
	PriorityQueueSize int

	// GenesisHeight is the height sampling will start from
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
//
// SamplingRange == 0 will cause the jobs' queue to be empty
//
//	therefore no sampling jobs will be reserved and more importantly
//	 the DASer will break
//
// ConcurrencyLimit == 0 will cause the number of workers to be 0 and thus
//
//	no threads will be assigned to the waiting jobs therefore breaking the DASer
//
// OR GenesisHeight == 0 there is no block height 0 to be sampled thus will break the DASer.
//
// # On the other hand
//
// BackgroundStoreInterval == 0 disables background storer,
// PriorityQueueSize == 0 disables prioritization of recently produced blocks for sampling
//
// Both of which won't break the DASer
func (cfg *Config) Validate() error {
	if !(cfg.SamplingRange > 0) {
		return errors.New("moddas misconfiguration: sampling range cannot be negative or 0")
	}

	if !(cfg.ConcurrencyLimit > 0) {
		return errors.New("moddas misconfiguration: concurrency limit cannot be negative or 0")
	}

	if !(cfg.GenesisHeight > 0) {
		return errors.New("moddas misconfiguration: genesis height cannot be negative or 0")
	}

	return nil
}
