package plugins

import "github.com/hootuu/utils/errors"

type Plugin interface {
	Init() *errors.Error
	StartUp() *errors.Error
}
