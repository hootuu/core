package mine

import (
	"github.com/hootuu/utils/sys"
	"github.com/ipfs/kubo/core"
	"sync"
)

type instance struct {
	node *core.IpfsNode

	lock sync.Mutex
}

var Mine = &instance{}

func (i *instance) BindNode(node *core.IpfsNode) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.node = node
}

func (i *instance) Node() *core.IpfsNode {
	if i.node == nil {
		sys.Error("must bind mine.node first.")
	}
	return i.node
}
