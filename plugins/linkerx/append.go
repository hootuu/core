package linkerx

import (
	"context"
	"github.com/hootuu/core/broadcast"
	"github.com/hootuu/core/hotu/here"
	"github.com/hootuu/domain/chain"
	"github.com/hootuu/linker"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/logger"
	"github.com/hootuu/utils/sys"
	"go.uber.org/zap"
)

const (
	AppendTopic = "linker.append"
)

type AppendListener struct {
}

func newAppendListener() *AppendListener {
	return &AppendListener{}
}

func (a *AppendListener) GetCode() string {
	return AppendTopic
}

func (a *AppendListener) Care(_ context.Context, msg *broadcast.Message) bool {
	if !here.Here.Mode().IsAngry() {
		return false
	}
	return true
}

func (a *AppendListener) Deal(_ context.Context, msg *broadcast.Message) *errors.Error {
	link, err := broadcast.MessageScan[chain.Link](msg)
	if err != nil {
		logger.Logger.Error("broadcast message scan failed", zap.Error(err))
		return err
	}
	bCare, err := linker.Care(link.GetCreation())
	if err != nil {
		logger.Logger.Error("linker care check failed", zap.Error(err))
		return err
	}
	if !bCare {
		return nil
	}
	lead, err := linker.Append(*link)
	if err != nil {
		logger.Logger.Error("linker care check failed", zap.Error(err))
		return err
	}
	if sys.RunMode.IsRd() {
		logger.Logger.Info("Write Linker", zap.Any("lead", lead))
	}
	return nil
}
