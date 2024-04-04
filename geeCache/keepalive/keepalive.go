package keepalive

const (
	KeepaliveStatusKey                 = "__keepalive_status"
	StatusOK           KeepaliveStatus = iota
)

type Keepalive interface {
	Keepalive()
}

type ResponseKeepalive interface {
	ResponseKeepalive()
}

type KeepaliveStatus = int64
