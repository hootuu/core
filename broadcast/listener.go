package broadcast

import "github.com/hootuu/utils/errors"

type Listener interface {
	GetCode() string
	Care(msg *Message) bool
	Deal(msg *Message) *errors.Error
}
