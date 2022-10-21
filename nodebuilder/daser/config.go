package daser

import (
	"errors"
	"time"
)

var (
	ErrNoSamplingRange = errors.New("nodebuilder/daser: misconfiguration error, sampling range cannot be 0")
	ErrNoConcLimit     = errors.New("nodebuikder/daser: misconfiguration error, concurrency limit cannot be 0")
	ErrNoGenesisHeight = errors.New("nodebuilder/daser: misconfiguration error, genesis height cannot be 0")
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
	// TODO(team): what should validate for in here? seems like all cfg.nfiguration fields can accept zero-values.
	if cfg.SamplingRange == 0 {
		return ErrNoSamplingRange
	}

	if cfg.ConcurrencyLimit == 0 {
		return ErrNoConcLimit
	}

	if cfg.GenesisHeight == 0 {
		return ErrNoGenesisHeight
	}

	return nil
}
