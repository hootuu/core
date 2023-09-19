package linkerx

import (
	"github.com/hootuu/core/broadcast"
	"github.com/hootuu/core/hotu/here"
	"github.com/hootuu/linker"
	"github.com/hootuu/utils/errors"
)

const (
	linkerDbPath = ".linker"
)

var LinkerX = &linkerX{}

type linkerX struct {
	AppendMQ *broadcast.MQ
}

func (l *linkerX) Init() *errors.Error {
	dbPath := here.Here.MustGetPath(linkerDbPath)
	err := linker.InitIfNeeded(dbPath)
	if err != nil {
		return err
	}
	l.AppendMQ, err = broadcast.NewMQ(AppendTopic)
	if err != nil {
		return err
	}
	return nil
}

func (l *linkerX) StartUp() *errors.Error {
	if here.Here.Mode().IsAngry() {
		l.AppendMQ.RegisterListener(newAppendListener())
		l.AppendMQ.StartListening()
	}
	return nil
}
