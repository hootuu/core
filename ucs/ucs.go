package ucs

import (
	"context"
	"fmt"
	"github.com/hootuu/core/hotu/here"
	"github.com/hootuu/utils/errors"
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
	"go.uber.org/zap"
)

const (
	ucsRepoFsPath = ".ucs"
)

var gUcsRepoFsPath string
var gUcsRepo repo.Repo
var gConfig *config.Config
var gPlugins *loader.PluginLoader
var gNode *core.IpfsNode

const (
	AlgorithmEd25519 = "ed25519"
)

func init() {
	gUcsRepoFsPath = here.Here.MustGetPath(ucsRepoFsPath)
}

func StartUp() *errors.Error {
	var nErr error
	_, nErr = loadPlugins()
	if nErr != nil {
		sys.Error("Startup UCS failed", nErr.Error())
		return errors.Sys("ucs load plugin failed")
	}
	if err := doInitIfNeeded(); err != nil {
		sys.Error("UCS Init failed", err.Error())
		return err
	}

	gUcsRepo, nErr = fsrepo.Open(gUcsRepoFsPath)
	if nErr != nil {
		sys.Error("Startup UCS failed", nErr.Error())
		return errors.Sys("fsrepo open failed")
	}
	gConfig, nErr = gUcsRepo.Config()
	if nErr != nil {
		sys.Error("Startup UCS failed", nErr.Error())
		return errors.Sys("ucs load config failed")
	}
	gNode, nErr = core.NewNode(context.Background(), &core.BuildCfg{
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
	here.Here.BindNode(gNode)

	var relayPeers []peer.AddrInfo

	if nErr := gNode.Bootstrap(bootstrap.BootstrapConfigWithPeers(relayPeers)); nErr != nil {
		gLogger.Error("ucs.node.Bootstrap failed", zap.Error(nErr))
		return errors.Sys("ucs node bootstrap failed")
	}
	return nil
}

func doInitIfNeeded() *errors.Error {
	if fsrepo.IsInitialized(gUcsRepoFsPath) {
		return nil
	}
	sys.Info("UCS Init At:", gUcsRepoFsPath)

	identity, nErr := config.CreateIdentity(
		logger.GetLoggerWriter(logger.Console),
		[]options.KeyGenerateOption{
			options.Key.Type(AlgorithmEd25519),
		},
	)
	if nErr != nil {
		sys.Error("ucs.CreateIdentity failed", nErr.Error())
		return errors.Sys("config CreateIdentity failed")
	}
	conf, nErr := config.InitWithIdentity(identity)
	if nErr != nil {
		sys.Error("usc Config InitWithIdentity failed", nErr.Error())
		return errors.Sys("config InitWithIdentity failed")
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

	if nErr := fsrepo.Init(gUcsRepoFsPath, conf); nErr != nil {
		sys.Error("UCS Init Failed:", nErr)
		return errors.Sys("fsrepo init failed")
	}
	return nil
}

func loadPlugins() (*loader.PluginLoader, error) {
	gPlugins, err := loader.NewPluginLoader(gUcsRepoFsPath)
	if err != nil {
		sys.Error("loader.NewPluginLoader failed", err.Error())
		return nil, fmt.Errorf("error loading plugins: %s", err)
	}
	if err := gPlugins.Initialize(); err != nil {
		sys.Error("plugins.Initialize failed", err.Error())
		return nil, fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := gPlugins.Inject(); err != nil {
		sys.Error("plugins.Inject failed", err.Error())
		return nil, fmt.Errorf("error inject plugins: %s", err)
	}
	return gPlugins, nil
}
