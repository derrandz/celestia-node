package das

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/fraud"
	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share"
)

var (
	log = logging.Logger("das")

	ErrInvalidOptionValue = func(optionName string, value string) error {
		errorMsg := fmt.Sprintf("das/daser: invalid option value: %s, cannot be %s", optionName, value)
		return errors.New(errorMsg)
	}
)

// optionSetter:
//
//	this is expected to be a closure that encloses the operation of parameter setting along side with
//	the value to be set.
//	example: both d and a are in the "global scope" relatively to the closure, thus no parameter passing is required.
//	   func() {
//
// .       d.params.option = a
//
//	}
type optionSetter func()
type Option func(*DASer) error
type OptionValidator func(value any, optionSetter optionSetter) error

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
func WithParamSamplingRange(samplingRange uint) Option {
	return func(d *DASer) error {
		return samplingRangeMustBeValid(samplingRange, func() {
			d.params.samplingRange = samplingRange
		})
	}
}

// OptionValidator for: samplingRange
func samplingRangeMustBeValid(samplingRange uint, setOption optionSetter) error {
	if samplingRange == 0 {
		return ErrInvalidOptionValue(
			"samplingRange",
			"0",
		)
	}

	setOption()
	return nil
}

// Option for: concurrencyLimit
func WithParamConcurrencyLimit(concurrencyLimit uint) Option {
	return func(d *DASer) error {
		return concurrencyLimitMustBeValid(concurrencyLimit, func() {
			d.params.concurrencyLimit = concurrencyLimit
		})
	}
}

// OptionValidator for: concurrencyLimit
func concurrencyLimitMustBeValid(concurrencyLimit uint, setOption optionSetter) error {
	if concurrencyLimit == 0 {
		return ErrInvalidOptionValue(
			"concurrencyLimit",
			"0",
		)
	}

	setOption()
	return nil
}

// Option for: bgStoreInterval
func WithParamBackgroundStoreInterval(bgStoreInterval time.Duration) Option {
	return func(d *DASer) error {
		return bgStoreIntervalMustBeValid(bgStoreInterval, func() {
			d.params.bgStoreInterval = bgStoreInterval
		})
	}
}

// OptionValidator for: bgStoreInterval
// No real validation is taking place because all values are valid.
// Using just to stay consistent with the pattern
// TODO(team): Maybe we should define an upper bound?
func bgStoreIntervalMustBeValid(bgStoreInterval time.Duration, setOption optionSetter) error {
	setOption()
	return nil
}

// Option for: priorityQueueSize
func WithParamPriorityQueueSize(priorityQueueSize uint) Option {
	return func(d *DASer) error {
		return priorityQueueSizeMustBeValid(priorityQueueSize, func() {
			d.params.priorityQueueSize = priorityQueueSize
		})
	}
}

// OptionValidaor for: priorityQueueSize
// No real validation is taking place because all values are valid.
// Using just to stay consistent with the pattern
func priorityQueueSizeMustBeValid(priorityQueueSize uint, setOption optionSetter) error {
	setOption()
	return nil
}

// Option for: genesisHeight
func WithParamGenesisHeight(genesisHeight uint) Option {
	return func(d *DASer) error {
		return genesisHeightMustBeValid(genesisHeight, func() {
			d.params.genesisHeight = genesisHeight
		})
	}
}

// OptionValidator for: genesisHeight
func genesisHeightMustBeValid(genesisHeight uint, setOption optionSetter) error {
	if genesisHeight == 0 {
		return ErrInvalidOptionValue(
			"genesisHeight",
			"0",
		)
	}

	setOption()
	return nil
}

// DASer continuously validates availability of data committed to headers.
type DASer struct {
	params parameters

	da     share.Availability
	bcast  fraud.Broadcaster
	hsub   header.Subscriber // listens for new headers in the network
	getter header.Getter     // retrieves past headers

	sampler    *samplingCoordinator
	store      checkpointStore
	subscriber subscriber

	cancel         context.CancelFunc
	subscriberDone chan struct{}
	running        int32
}

type listenFn func(ctx context.Context, height uint64)
type sampleFn func(context.Context, *header.ExtendedHeader) error

// NewDASer creates a new DASer.
func NewDASer(
	da share.Availability,
	hsub header.Subscriber,
	getter header.Getter,
	dstore datastore.Datastore,
	bcast fraud.Broadcaster,
	options ...Option,
) *DASer {
	d := &DASer{
		params:         defaultParameters(),
		da:             da,
		bcast:          bcast,
		hsub:           hsub,
		getter:         getter,
		store:          newCheckpointStore(dstore),
		subscriber:     newSubscriber(),
		subscriberDone: make(chan struct{}),
	}

	for _, applyOpt := range options {
		err := applyOpt(d)
		if err != nil {
			panic(err)
		}
	}

	d.sampler = newSamplingCoordinator(d.params, getter, d.sample)

	return d
}

// Start initiates subscription for new ExtendedHeaders and spawns a sampling routine.
func (d *DASer) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&d.running, 0, 1) {
		return fmt.Errorf("da: DASer already started")
	}

	sub, err := d.hsub.Subscribe()
	if err != nil {
		return err
	}

	// load latest DASed checkpoint
	cp, err := d.store.load(ctx)
	if err != nil {
		log.Warnw("checkpoint not found, initializing with height 1")

		cp = checkpoint{
			SampleFrom:  uint64(d.params.genesisHeight),
			NetworkHead: uint64(d.params.genesisHeight),
		}

		// attempt to get head info. No need to handle error, later DASer
		// will be able to find new head from subscriber after it is started
		if h, err := d.getter.Head(ctx); err == nil {
			cp.NetworkHead = uint64(h.Height)
		}
	}
	log.Info("starting DASer from checkpoint: ", cp.String())

	runCtx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel

	go d.sampler.run(runCtx, cp)
	go d.subscriber.run(runCtx, sub, d.sampler.listen)
	go d.store.runBackgroundStore(runCtx, d.params.bgStoreInterval, d.sampler.getCheckpoint)

	return nil
}

// Stop stops sampling.
func (d *DASer) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&d.running, 1, 0) {
		return nil
	}

	// try to store checkpoint without waiting for coordinator and workers to stop
	cp, err := d.sampler.getCheckpoint(ctx)
	if err != nil {
		log.Error("DASer coordinator checkpoint is unavailable")
	}

	if err = d.store.store(ctx, cp); err != nil {
		log.Errorw("storing checkpoint to disk", "err", err)
	}

	d.cancel()
	if err = d.sampler.wait(ctx); err != nil {
		return fmt.Errorf("DASer force quit: %w", err)
	}

	// save updated checkpoint after sampler and all workers are shut down
	if err = d.store.store(ctx, newCheckpoint(d.sampler.state.unsafeStats())); err != nil {
		log.Errorw("storing checkpoint to disk", "Err", err)
	}

	if err = d.store.wait(ctx); err != nil {
		return fmt.Errorf("DASer force quit with err: %w", err)
	}
	return d.subscriber.wait(ctx)
}

func (d *DASer) sample(ctx context.Context, h *header.ExtendedHeader) error {
	err := d.da.SharesAvailable(ctx, h.DAH)
	if err != nil {
		if err == context.Canceled {
			return err
		}
		var byzantineErr *share.ErrByzantine
		if errors.As(err, &byzantineErr) {
			log.Warn("Propagating proof...")
			sendErr := d.bcast.Broadcast(ctx, fraud.CreateBadEncodingProof(h.Hash(), uint64(h.Height), byzantineErr))
			if sendErr != nil {
				log.Errorw("fraud proof propagating failed", "err", sendErr)
			}
		}

		log.Errorw("sampling failed", "height", h.Height, "hash", h.Hash(),
			"square width", len(h.DAH.RowsRoots), "data root", h.DAH.Hash(), "err", err)
		return err
	}

	return nil
}

func (d *DASer) SamplingStats(ctx context.Context) (SamplingStats, error) {
	return d.sampler.stats(ctx)
}
