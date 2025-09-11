package p2p

import "net"

// This interface represents a remote node
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Transport interface handles the communation between the nodes in the network
// It can be of type TCP, UPD, WEBSOCKETS, ...
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
