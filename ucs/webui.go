package ucs

import (
	"fmt"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	cmds "github.com/ipfs/go-ipfs-cmds"
	cmdshttp "github.com/ipfs/go-ipfs-cmds/http"
	version "github.com/ipfs/kubo"
	oldcmds "github.com/ipfs/kubo/commands"
	"github.com/ipfs/kubo/core"
	corecommands "github.com/ipfs/kubo/core/commands"
	"github.com/ipfs/kubo/core/corehttp"
	sockets "github.com/libp2p/go-socket-activation"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func StartWebui() *errors.Error {
	listeners, nErr := sockets.TakeListeners("io.ipfs.api")
	if nErr != nil {
		gLogger.Error("sockets.TakeListeners failed", zap.Error(nErr))
		return errors.Sys("sockets.take.listener failed", nErr)
	}

	cfg, nErr := gUcsRepo.Config()
	if nErr != nil {
		gLogger.Error("gUcsRepo.Config failed", zap.Error(nErr))
		return errors.Sys("load config failed", nErr)
	}
	apiAddrs := cfg.Addresses.API

	listenerAddrs := make(map[string]bool, len(listeners))
	for _, listener := range listeners {
		listenerAddrs[string(listener.Multiaddr().Bytes())] = true
	}

	for _, addr := range apiAddrs {
		apiMaddr, nErr := ma.NewMultiaddr(addr)
		if nErr != nil {
			gLogger.Info("ma.NewMultiaddr failed", zap.Error(nErr))
			return errors.Sys("ma.NewMultiaddr failed")
		}
		if listenerAddrs[string(apiMaddr.Bytes())] {
			continue
		}

		apiLis, nErr := manet.Listen(apiMaddr)
		if nErr != nil {
			gLogger.Error("manet.Listen failed", zap.Error(nErr))
			return errors.Sys("manet.Listen failed")
		}

		listenerAddrs[string(apiMaddr.Bytes())] = true
		listeners = append(listeners, apiLis)
	}

	for _, listener := range listeners {
		sys.Info("ucs.rpc.api listening on", listener.Multiaddr())
		switch listener.Addr().Network() {
		case "tcp", "tcp4", "tcp6":
			sys.Info("ucs.webui: http://", listener.Addr(), "/webui")
		}
	}

	gatewayOpt := corehttp.GatewayOption(corehttp.WebUIPaths...)
	//gatewayOpt = corehttp.GatewayOption("/ipfs", "/ipns")

	opts := []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("api"),
		corehttp.MetricsOpenCensusCollectionOption(),
		corehttp.MetricsOpenCensusDefaultPrometheusRegistry(),
		corehttp.CheckVersionOption(),
		commandsOption(oldcmds.Context{
			ConfigRoot: gUcsRepoFsPath,
			ReqLog:     &oldcmds.ReqLog{},
			Plugins:    gPlugins,
			Gateway:    true,
			ConstructNode: func() (*core.IpfsNode, error) {
				return gNode, nil
			},
		}, corecommands.Root, true),
		corehttp.WebUIOption,
		gatewayOpt,
		corehttp.VersionOption(),
		corehttp.MutexFractionOption("/debug/pprof-mutex/"),
		corehttp.BlockProfileRateOption("/debug/pprof-block/"),
		corehttp.MetricsScrapingOption("/debug/metrics/prometheus"),
		corehttp.LogOption(),
	}

	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	if nErr := gUcsRepo.SetAPIAddr(rewriteMaddrToUseLocalhostIfItsAny(listeners[0].Multiaddr())); nErr != nil {
		gLogger.Error("gUcsRepo.SetAPIAddr error", zap.Error(nErr))
		return errors.Sys("SetAPIAddr failed")
	}

	for _, apiLis := range listeners {
		go func(lis manet.Listener) {
			nErr := corehttp.Serve(gNode, manet.NetListener(lis), opts...)
			if nErr != nil {
				gLogger.Error("corehttp.Serve error", zap.Error(nErr))
			}
		}(apiLis)
	}

	fmt.Println("gogogog....")

	return nil
}

func commandsOption(cctx oldcmds.Context, command *cmds.Command, allowGet bool) corehttp.ServeOption {
	return func(n *core.IpfsNode, l net.Listener, mux *http.ServeMux) (*http.ServeMux, error) {

		cfg := cmdshttp.NewServerConfig()
		cfg.AllowGet = allowGet
		corsAllowedMethods := []string{http.MethodPost}
		if allowGet {
			corsAllowedMethods = append(corsAllowedMethods, http.MethodGet)
		}

		cfg.SetAllowedMethods(corsAllowedMethods...)
		cfg.APIPath = "/api/v0"

		addHeadersFromConfig(cfg)
		addCORSFromEnv(cfg)
		addCORSDefaults(cfg)
		patchCORSVars(cfg, l.Addr())

		cmdHandler := cmdshttp.NewHandler(&cctx, command, cfg)
		cmdHandler = otelhttp.NewHandler(cmdHandler, "corehttp.cmdsHandler")
		mux.Handle(cfg.APIPath+"/", cmdHandler)
		return mux, nil
	}
}

func addCORSFromEnv(c *cmdshttp.ServerConfig) {
	origin := os.Getenv("API_ORIGIN")
	if origin != "" {
		//log.Warn(originEnvKeyDeprecate)
		c.AppendAllowedOrigins(origin)
	}
}

func addHeadersFromConfig(c *cmdshttp.ServerConfig) {
	fmt.Println("Using API.HTTPHeaders:", gConfig.API.HTTPHeaders)

	if acao := gConfig.API.HTTPHeaders[cmdshttp.ACAOrigin]; acao != nil {
		c.SetAllowedOrigins(acao...)
	}
	if acam := gConfig.API.HTTPHeaders[cmdshttp.ACAMethods]; acam != nil {
		c.SetAllowedMethods(acam...)
	}
	for _, v := range gConfig.API.HTTPHeaders[cmdshttp.ACACredentials] {
		c.SetAllowCredentials(strings.ToLower(v) == "true")
	}

	c.Headers = make(map[string][]string, len(gConfig.API.HTTPHeaders)+1)

	// Copy these because the config is shared and this function is called
	// in multiple places concurrently. Updating these in-place *is* racy.
	for h, v := range gConfig.API.HTTPHeaders {
		h = http.CanonicalHeaderKey(h)
		switch h {
		case cmdshttp.ACAOrigin, cmdshttp.ACAMethods, cmdshttp.ACACredentials:
			// these are handled by the CORs library.
		default:
			c.Headers[h] = v
		}
	}
	c.Headers["Server"] = []string{"kubo/" + version.CurrentVersionNumber}
}

var defaultLocalhostOrigins = []string{
	"http://127.0.0.1:<port>",
	"https://127.0.0.1:<port>",
	"http://[::1]:<port>",
	"https://[::1]:<port>",
	"http://localhost:<port>",
	"https://localhost:<port>",
}

var companionBrowserExtensionOrigins = []string{
	"chrome-extension://nibjojkomfdiaoajekhjakgkdhaomnch", // ipfs-companion
	"chrome-extension://hjoieblefckbooibpepigmacodalfndh", // ipfs-companion-beta
}

func addCORSDefaults(c *cmdshttp.ServerConfig) {
	// always safelist certain origins
	c.AppendAllowedOrigins(defaultLocalhostOrigins...)
	c.AppendAllowedOrigins(companionBrowserExtensionOrigins...)

	// by default, use GET, PUT, POST
	if len(c.AllowedMethods()) == 0 {
		c.SetAllowedMethods(http.MethodGet, http.MethodPost, http.MethodPut)
	}
}

func patchCORSVars(c *cmdshttp.ServerConfig, addr net.Addr) {

	// we have to grab the port from an addr, which may be an ip6 addr.
	// TODO: this should take multiaddrs and derive port from there.
	port := ""
	if tcpaddr, ok := addr.(*net.TCPAddr); ok {
		port = strconv.Itoa(tcpaddr.Port)
	} else if udpaddr, ok := addr.(*net.UDPAddr); ok {
		port = strconv.Itoa(udpaddr.Port)
	}

	// we're listening on tcp/udp with ports. ("udp!?" you say? yeah... it happens...)
	oldOrigins := c.AllowedOrigins()
	newOrigins := make([]string, len(oldOrigins))
	for i, o := range oldOrigins {
		// TODO: allow replacing <host>. tricky, ip4 and ip6 and hostnames...
		if port != "" {
			o = strings.Replace(o, "<port>", port, -1)
		}
		newOrigins[i] = o
	}
	c.SetAllowedOrigins(newOrigins...)
}
