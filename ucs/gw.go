package ucs

import (
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"github.com/ipfs/kubo/core/corehttp"
	sockets "github.com/libp2p/go-socket-activation"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"go.uber.org/zap"
)

func StartGW() *errors.Error {

	listeners, nErr := sockets.TakeListeners("io.ipfs.gateway")
	if nErr != nil {
		gLogger.Error("sockets.TakeListeners failed", zap.Error(nErr))
		return errors.Sys("StartGW Failed:" + nErr.Error())
	}

	listenerAddrMap := make(map[string]bool, len(listeners))
	for _, listener := range listeners {
		listenerAddrMap[string(listener.Multiaddr().Bytes())] = true
	}

	cfg, nErr := gUcsRepo.Config()
	if nErr != nil {
		gLogger.Error("gUcsRepo.Config failed", zap.Error(nErr))
		return errors.Sys("StartGW Failed:" + nErr.Error())
	}

	gatewayAddrArr := cfg.Addresses.Gateway
	if len(gatewayAddrArr) == 0 {
		return nil
	}
	for _, addr := range gatewayAddrArr {
		gatewayMaddr, nErr := ma.NewMultiaddr(addr)
		if nErr != nil {
			gLogger.Error("ma.NewMultiaddr failed", zap.Error(nErr))
			return errors.Sys("StartGW Failed:" + nErr.Error())
		}

		if listenerAddrMap[string(gatewayMaddr.Bytes())] {
			continue
		}

		gwLis, nErr := manet.Listen(gatewayMaddr)
		if nErr != nil {
			gLogger.Error("manet.Listen failed", zap.Error(nErr))
			return errors.Sys("StartGW Failed:" + nErr.Error())
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
		addr, nErr := manet.ToNetAddr(rewriteMaddrToUseLocalhostIfItsAny(listeners[0].Multiaddr()))
		if nErr != nil {
			gLogger.Error("manet.ToNetAddr failed", zap.Error(nErr))
			return errors.Sys("StartGW Failed:" + nErr.Error())
		}
		if nErr := gUcsRepo.SetGatewayAddr(addr); nErr != nil {
			gLogger.Error("gUcsRepo.SetGatewayAddr failed", zap.Error(nErr))
			return errors.Sys("StartGW Failed:" + nErr.Error())
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
