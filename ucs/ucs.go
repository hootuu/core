package ucs

import (
	"context"
	"fmt"
	"github.com/hootuu/utils/logger"
	"github.com/hootuu/utils/sys"
	"github.com/ipfs/boxo/coreiface/options"
	"github.com/ipfs/kubo/config"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/bootstrap"
	"github.com/ipfs/kubo/core/node/libp2p"
	"github.com/ipfs/kubo/plugin/loader"
	"github.com/ipfs/kubo/repo"
	"github.com/ipfs/kubo/repo/fsrepo"
	"github.com/libp2p/go-libp2p/core/peer"
	"github/hootuu/core/mine"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
)

var gUcsRepoFsPath = "./.hotu/ucs"
var gUcsRepo repo.Repo
var gConfig *config.Config
var gPlugins *loader.PluginLoader
var gNode *core.IpfsNode

const (
	AlgorithmEd25519 = "ed25519"
)

func StartUp() error {
	_, err := loadPlugins()
	if err != nil {
		gLogger.Error("ucs.load.plugins failed", zap.Error(err))
		return err
	}
	err = doInitIfNeeded()
	if err != nil {
		gLogger.Error("usc.init failed", zap.Error(err))
		return err
	}

	gUcsRepo, err = fsrepo.Open(gUcsRepoFsPath)
	if err != nil {
		gLogger.Error("ucs.fs.Open failed", zap.Error(err))
		return err
	}
	gConfig, err = gUcsRepo.Config()
	if err != nil {
		gLogger.Error("ucs.fs.Open failed", zap.Error(err))
		return err
	}
	gNode, err = core.NewNode(context.Background(), &core.BuildCfg{
		Online:                      true,
		Repo:                        gUcsRepo,
		Permanent:                   true,
		DisableEncryptedConnections: false,
		Routing:                     libp2p.DHTOption,
		ExtraOpts: map[string]bool{
			"pubsub": true,
			"ipnsps": true,
		},
	})
	gNode.IsDaemon = true
	mine.Mine.BindNode(gNode)
	sys.Info("ucs node id: ", mine.Mine.Node().Identity.String())

	var relayPeers []peer.AddrInfo

	if err := gNode.Bootstrap(bootstrap.BootstrapConfigWithPeers(relayPeers)); err != nil {
		slog.Error("ucs.node.Bootstrap failed", err)
		return err
	}
	return nil
}

func doInitIfNeeded() error {
	if fsrepo.IsInitialized(gUcsRepoFsPath) {
		return nil
	}
	slog.Info("init hotu-ucs at:", gUcsRepoFsPath)

	identity, err := config.CreateIdentity(
		logger.GetLoggerWriter(logger.Console),
		[]options.KeyGenerateOption{
			options.Key.Type(AlgorithmEd25519),
		},
	)
	if err != nil {
		slog.Error("ucs.CreateIdentity failed", err)
		return err
	}
	conf, err := config.InitWithIdentity(identity)
	if err != nil {
		slog.Error("usc Config InitWithIdentity failed", err)
		return err
	}
	//conf.API.HTTPHeaders = make(map[string][]string)
	//conf.API.HTTPHeaders["Access-Control-Allow-Origin"] = []string{"*"}
	//conf.API.HTTPHeaders["Access-Control-Allow-Methods"] = []string{"GET", "POST"}

	//conf.Datastore = config.Datastore{
	//	StorageMax:         "10GB",
	//	StorageGCWatermark: 90, // 90%
	//	GCPeriod:           "1h",
	//	BloomFilterSize:    0,
	//	Spec: map[string]interface{}{
	//		"type":   "measure",
	//		"prefix": "badger.datastore",
	//		"child": map[string]interface{}{
	//			"type":       "badgerds",
	//			"path":       "badgerds",
	//			"syncWrites": false,
	//			"truncate":   true,
	//		},
	//	},
	//}
	//conf.Experimental = config.Experiments{
	//	FilestoreEnabled:              true,
	//	UrlstoreEnabled:               true,
	//	ShardingEnabled:               false,
	//	GraphsyncEnabled:              false,
	//	Libp2pStreamMounting:          false,
	//	P2pHttpProxy:                  true,
	//	StrategicProviding:            false,
	//	AcceleratedDHTClient:          false,
	//	OptimisticProvide:             false,
	//	OptimisticProvideJobsPoolSize: 10,
	//}

	if err := fsrepo.Init(gUcsRepoFsPath, conf); err != nil {
		slog.Error("ucs.fs.Init failed", err)
		return err
	}
	return nil
}

func loadPlugins() (*loader.PluginLoader, error) {
	gPlugins, err := loader.NewPluginLoader(gUcsRepoFsPath)
	if err != nil {
		slog.Error("loader.NewPluginLoader failed", err)
		return nil, fmt.Errorf("error loading plugins: %s", err)
	}
	if err := gPlugins.Initialize(); err != nil {
		slog.Error("plugins.Initialize failed", err)
		return nil, fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := gPlugins.Inject(); err != nil {
		slog.Error("plugins.Inject failed", err)
		return nil, fmt.Errorf("error inject plugins: %s", err)
	}
	return gPlugins, nil
}
