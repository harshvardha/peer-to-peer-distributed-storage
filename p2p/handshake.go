package p2p

// HandshakeFunc is
type HandeshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error { return nil }
