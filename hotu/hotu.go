package hotu

import "github.com/hootuu/domain/scope"

type Mode int32

func (m Mode) IsVN() bool {
	return m == vnMode
}

func (m Mode) IsScope() bool {
	return m == scopeMode
}

func (m Mode) IsPeer() bool {
	return m == peerMode
}

const (
	vnMode    Mode = 9999
	scopeMode Mode = 8888
	peerMode  Mode = 6666
)

type modes struct {
	VN    Mode
	Scope Mode
	Peer  Mode
}

var Modes = modes{
	VN:    vnMode,
	Scope: scopeMode,
	Peer:  peerMode,
}

type Hotu struct {
	Mode Mode
	Lead scope.Lead
}

func (h *Hotu) Init() {

}

var HOTU *Hotu
