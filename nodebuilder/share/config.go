package share

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNegativeInterval = errors.New("interval must be positive")
)

type Config struct {
	// PeersLimit defines how many peers will be added during discovery.
	PeersLimit uint
	// DiscoveryInterval is an interval between discovery sessions.
	DiscoveryInterval time.Duration
	// AdvertiseInterval is a interval between advertising sessions.
	// NOTE: only full and bridge can advertise themselves.
	AdvertiseInterval time.Duration

	// Used to disable usage of cascade in favor of configured default getter
	NoCascade bool

	// DefaultGetter is a default getter to be used in case of no cascade
	DefaultGetter string

	// Used to enable/disable the use of IPLD fallback for the retrieval of data.
	UseIPLDFallback bool
}

func DefaultConfig() Config {
	return Config{
		PeersLimit:        3,
		DiscoveryInterval: time.Second * 30,
		AdvertiseInterval: time.Second * 30,
		UseIPLDFallback:   true,
		NoCascade:         false,
	}
}

// Validate performs basic validation of the config.
func (cfg *Config) Validate() error {
	if cfg.DiscoveryInterval <= 0 || cfg.AdvertiseInterval <= 0 {
		return fmt.Errorf("nodebuilder/share: %s", ErrNegativeInterval)
	}

	if cfg.NoCascade == true && cfg.DefaultGetter == "" {
		return errors.New("nodebuilder/share: no default getter provided")
	}

	return nil
}
