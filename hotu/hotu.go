package hotu

import (
	"github.com/hootuu/core/hotu/here"
	"github.com/hootuu/core/plugins/linkerx"
	"github.com/hootuu/core/ucs"
	"github.com/hootuu/domain/point"
	"github.com/hootuu/domain/scope"
	"github.com/hootuu/utils/logger"
	"github.com/hootuu/utils/sys"
	"github.com/ipfs/kubo/core"
	"go.uber.org/zap"
)

type ID string

func (id ID) Equals(oId string) bool {
	return string(id) == oId
}

type hotu struct {
	_initialized bool
}

func (h *hotu) Mode() point.Mode {
	return here.Here.Mode()
}

func (h *hotu) ID() ID {
	return ID(here.Here.ID())
}

func (h *hotu) Node() *core.IpfsNode {
	return here.Here.Node()
}

func (h *hotu) SetMode(mode point.Mode) {
	here.Here.SetMode(mode)
}

func (h *hotu) SetLead(lead scope.Lead) {
	here.Here.SetLead(lead)
}

func (h *hotu) Init(mode point.Mode, lead scope.Lead) {
	if h._initialized {
		return
	}
	h.SetMode(mode)
	h.SetLead(lead)
	if err := ucs.StartUp(); err != nil {
		sys.Error("ucs.StartUp failed", zap.Error(err))
		sys.Exit(err)
		return
	}
	if err := linkerx.LinkerX.Init(); err != nil {
		sys.Error("LinkerX init failed", zap.Error(err))
		sys.Exit(err)
		return
	}
	h._initialized = true
}

func (h *hotu) StartUp() {
	if err := linkerx.LinkerX.StartUp(); err != nil {
		sys.Error("LinkerX.StartUp failed", zap.Error(err))
		sys.Exit(err)
		return
	}

	if err := ucs.StartGW(); err != nil {
		sys.Error("ucs.StartGW failed", zap.Error(err))
		sys.Exit(err)
		return
	}

	if err := ucs.StartWebui(); err != nil {
		logger.Logger.Info("ucs.StartWebui failed", zap.String("err", err.Error()))
		sys.Exit(err)
		return
	}
}

var Hotu = &hotu{}
