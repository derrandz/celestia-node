package share

import (
	"context"

	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/libs/fxutil"
	libhead "github.com/celestiaorg/celestia-node/libs/header"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	modp2p "github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/share"
	disc "github.com/celestiaorg/celestia-node/share/availability/discovery"
	"github.com/celestiaorg/celestia-node/share/availability/full"
	"github.com/celestiaorg/celestia-node/share/availability/light"
	"github.com/celestiaorg/celestia-node/share/eds"
	"github.com/celestiaorg/celestia-node/share/getters"
	"github.com/celestiaorg/celestia-node/share/p2p/peers"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexeds"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexnd"
	"github.com/celestiaorg/celestia-node/share/p2p/shrexsub"
)

var log = logging.Logger("nodebuilder/share")

func ConstructModule(tp node.Type, cfg *Config, options ...fx.Option) fx.Option {
	// sanitize config values before constructing module
	cfgErr := cfg.Validate()

	baseComponents := fx.Options(
		fx.Supply(*cfg),
		fx.Error(cfgErr),
		fx.Options(options...),
		fx.Provide(discovery(*cfg)),
		fx.Provide(newModule),
		fx.Invoke(share.EnsureEmptySquareExists),
		// TODO: Configure for light nodes
		fx.Provide(
			func(host host.Host, network modp2p.Network) (*shrexnd.Client, error) {
				return shrexnd.NewClient(host, shrexnd.WithProtocolSuffix(string(network)))
			},
		),
	)

	sharedBridgeFullComponents := fx.Options(
		baseComponents,
		fx.Provide(getters.NewIPLDGetter),
		fx.Invoke(func(srv *shrexeds.Server) {}),
		fx.Provide(fx.Annotate(
			func(host host.Host, store *eds.Store, network modp2p.Network) (*shrexeds.Server, error) {
				return shrexeds.NewServer(host, store, shrexeds.WithProtocolSuffix(string(network)))
			},
			fx.OnStart(func(ctx context.Context, server *shrexeds.Server) error {
				return server.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, server *shrexeds.Server) error {
				return server.Stop(ctx)
			}),
		)),
		// Bridge Nodes need a client as well, for requests over FullAvailability
		fx.Provide(
			func(host host.Host, network modp2p.Network) (*shrexeds.Client, error) {
				return shrexeds.NewClient(host, shrexeds.WithProtocolSuffix(string(network)))
			},
		),
		fx.Provide(fx.Annotate(
			func(path node.StorePath, ds datastore.Batching) (*eds.Store, error) {
				return eds.NewStore(string(path), ds)
			},
			fx.OnStart(func(ctx context.Context, store *eds.Store) error {
				err := store.Start(ctx)
				if err != nil {
					return err
				}

				return ensureEmptyCARExists(ctx, store)
			}),
			fx.OnStop(func(ctx context.Context, store *eds.Store) error {
				return store.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			func(ctx context.Context, h host.Host, network modp2p.Network) (*shrexsub.PubSub, error) {
				return shrexsub.NewPubSub(
					ctx,
					h,
					string(network),
				)
			},
			fx.OnStart(func(ctx context.Context, pubsub *shrexsub.PubSub) error {
				return pubsub.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, pubsub *shrexsub.PubSub) error {
				err := pubsub.Stop(ctx)
				if err != nil {
					// TODO: fix lifecycle from shrexgetter to avoid not being able to return this err
					log.Warnf("stopping shrexsub: %v", err)
				}
				return nil
			}),
		)),
		fx.Provide(fx.Annotate(
			full.NewShareAvailability,
			fx.OnStart(func(ctx context.Context, avail *full.ShareAvailability) error {
				return avail.Start(ctx)
			}),
			fx.OnStop(func(ctx context.Context, avail *full.ShareAvailability) error {
				return avail.Stop(ctx)
			}),
		)),
		// cacheAvailability's lifecycle continues to use a fx hook,
		// since the LC requires a cacheAvailability but the constructor returns a share.Availability
		fx.Provide(cacheAvailability[*full.ShareAvailability]),
	)

	switch tp {
	case node.Light:
		return fx.Module(
			"share",
			baseComponents,
			fxutil.ProvideAs(getters.NewIPLDGetter, new(share.Getter)),
			fx.Provide(fx.Annotate(light.NewShareAvailability)),
			// cacheAvailability's lifecycle continues to use a fx hook,
			// since the LC requires a cacheAvailability but the constructor returns a share.Availability
			fx.Provide(cacheAvailability[*light.ShareAvailability]),
		)
	case node.Full:
		return fx.Module(
			"share",
			sharedBridgeFullComponents,
			fx.Provide(fx.Annotate(
				getters.NewShrexGetter,
				fx.OnStart(func(ctx context.Context, getter *getters.ShrexGetter) error {
					return getter.Start(ctx)
				}),
				fx.OnStop(func(ctx context.Context, getter *getters.ShrexGetter) error {
					return getter.Stop(ctx)
				}),
			)),
			fx.Provide(fullGetter),
			fx.Provide(peerManager),
		)
	case node.Bridge:
		return fx.Module(
			"share",
			sharedBridgeFullComponents,
			fx.Provide(bridgeGetter),
		)
	default:
		panic("invalid node type")
	}
}

func peerManager(subscriber libhead.Subscriber[*header.ExtendedHeader], host host.Host, discovery *disc.Discovery) *peers.Manager {
	// TODO: Replace modp2p.BlockTime?
	return peers.NewManager(subscriber, host, discovery, modp2p.BlockTime)
}

func bridgeGetter(
	store *eds.Store,
	ipldGetter *getters.IPLDGetter,
) share.Getter {
	return getters.NewCascadeGetter(
		[]share.Getter{
			getters.NewStoreGetter(store),
			getters.NewTeeGetter(ipldGetter, store),
		},
		// TODO: Replace modp2p.BlockTime?
		modp2p.BlockTime,
	)
}

// TODO: Light nodes should also use shrexgetter for nd
func fullGetter(
	store *eds.Store,
	shrexGetter *getters.ShrexGetter,
	ipldGetter *getters.IPLDGetter,
) share.Getter {
	return getters.NewCascadeGetter(
		[]share.Getter{
			getters.NewStoreGetter(store),
			getters.NewTeeGetter(shrexGetter, store),
			// getters.NewTeeGetter(ipldGetter, store),
		},
		// TODO: Replace modp2p.BlockTime?
		modp2p.BlockTime,
	)
}
