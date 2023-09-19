package hotu

type Mode int32

func (m Mode) IsAngry() bool {
	return m == angryMode
}

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
	angryMode Mode = 1010
	vnMode    Mode = 9999
	scopeMode Mode = 8888
	peerMode  Mode = 6666
)

type modes struct {
	Angry Mode
	VN    Mode
	Scope Mode
	Peer  Mode
}

var Modes = modes{
	Angry: angryMode,
	VN:    vnMode,
	Scope: scopeMode,
	Peer:  peerMode,
}
