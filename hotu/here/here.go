package here

import (
	"github.com/hootuu/domain/point"
	"github.com/hootuu/domain/scope"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"github.com/ipfs/kubo/core"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

const (
	UcsHomePath = ".hotu"
)

type here struct {
	homeDir string
	mode    point.Mode
	lead    scope.Lead
	id      string
	node    *core.IpfsNode

	lock sync.Mutex
}

var Here *here

func (i *here) SetMode(mode point.Mode) {
	i.mode = mode
}

func (i *here) SetLead(lead scope.Lead) {
	i.lead = lead
}

func (i *here) BindNode(node *core.IpfsNode) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.node = node
	i.id = i.node.Identity.String()
}

func (i *here) MustGetPath(subPath string) string {
	path := filepath.Join(i.homeDir, subPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			sys.Error("Create Path Failed: ", path)
			os.Exit(-1)
		}
	}
	return path
}

func (i *here) Mode() point.Mode {
	return i.mode
}

func (i *here) Lead() scope.Lead {
	return i.lead
}

func (i *here) ID() string {
	return i.id
}

func (i *here) Node() *core.IpfsNode {
	return i.node
}

func init() {
	Here = &here{}
	//userHomeDir, err := getUserHomeDir()
	//if err != nil {
	//	os.Exit(-1)
	//}
	Here.homeDir = filepath.Join(".", UcsHomePath)
}

func getUserHomeDir() (string, *errors.Error) {
	currentUser, nErr := user.Current()
	if nErr != nil {
		sys.Error("Cannot get the current user")
		return "", errors.Sys("can not get the current user")
	}
	sys.Info("Current User Home Directory:", currentUser.HomeDir)
	return currentUser.HomeDir, nil
}
