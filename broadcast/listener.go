package broadcast

import (
	"context"
	"github.com/hootuu/utils/errors"
)

type Listener interface {
	GetCode() string
	Care(ctx context.Context, msg *Message) bool
	Deal(ctx context.Context, msg *Message) *errors.Error
}
