package p2p

// This interface represents a remote node
type Peer interface {
}

// Transport interface handles the communation between the nodes in the network
// It can be of type TCP, UPD, WEBSOCKETS, ...
type Transport interface {
	ListenAndAccept() error
}
