package ucs

import (
	"github.com/hootuu/utils/sys"
	"github.com/ipfs/kubo/core/corehttp"
	sockets "github.com/libp2p/go-socket-activation"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/zap"
)

func StartGW() error {

	listeners, err := sockets.TakeListeners("io.ipfs.gateway")
	if err != nil {
		gLogger.Error("sockets.TakeListeners failed", zap.Error(err))
		return err
	}

	listenerAddrMap := make(map[string]bool, len(listeners))
	for _, listener := range listeners {
		listenerAddrMap[string(listener.Multiaddr().Bytes())] = true
	}

	cfg, err := gUcsRepo.Config()
	if err != nil {
		gLogger.Error("gUcsRepo.Config failed", zap.Error(err))
		return err
	}

	gatewayAddrArr := cfg.Addresses.Gateway
	if len(gatewayAddrArr) == 0 {
		return nil
	}
	for _, addr := range gatewayAddrArr {
		gatewayMaddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			gLogger.Error("ma.NewMultiaddr failed", zap.Error(err))
			return err
		}

		if listenerAddrMap[string(gatewayMaddr.Bytes())] {
			continue
		}

		gwLis, err := manet.Listen(gatewayMaddr)
		if err != nil {
			gLogger.Error("manet.Listen failed", zap.Error(err))
			return err
		}
		listenerAddrMap[string(gatewayMaddr.Bytes())] = true
		listeners = append(listeners, gwLis)
	}

	for _, listener := range listeners {
		sys.Info("Gateway server listening on", listener.Multiaddr())
	}

	opts := []corehttp.ServeOption{
		corehttp.MetricsCollectionOption("gateway"),
		corehttp.HostnameOption(),
		corehttp.GatewayOption("/ipfs", "/ipns"),
		corehttp.VersionOption(),
		corehttp.CheckVersionOption(),
	}

	if cfg.Experimental.P2pHttpProxy {
		opts = append(opts, corehttp.P2PProxyOption())
	}
	if len(cfg.Gateway.RootRedirect) > 0 {
		opts = append(opts, corehttp.RedirectOption("", cfg.Gateway.RootRedirect))
	}

	if len(listeners) > 0 {
		addr, err := manet.ToNetAddr(rewriteMaddrToUseLocalhostIfItsAny(listeners[0].Multiaddr()))
		if err != nil {
			gLogger.Error("manet.ToNetAddr failed", zap.Error(err))
			return err
		}
		if err := gUcsRepo.SetGatewayAddr(addr); err != nil {
			gLogger.Error("gUcsRepo.SetGatewayAddr failed", zap.Error(err))
			return err
		}
	}

	for _, lis := range listeners {
		go func(lis manet.Listener) {
			err := corehttp.Serve(gNode, manet.NetListener(lis), opts...)
			if err != nil {
				gLogger.Error("corehttp.Serve failed", zap.Error(err))
			}
		}(lis)
	}

	return nil
}

func rewriteMaddrToUseLocalhostIfItsAny(maddr ma.Multiaddr) ma.Multiaddr {
	first, rest := ma.SplitFirst(maddr)

	switch {
	case first.Equal(manet.IP4Unspecified):
		return manet.IP4Loopback.Encapsulate(rest)
	case first.Equal(manet.IP6Unspecified):
		return manet.IP6Loopback.Encapsulate(rest)
	default:
		return maddr // not ip
	}
}
