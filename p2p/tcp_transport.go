package p2p

import (
	"log"
	"net"
	"sync"
)

// It represents the remote node using TCP as transport layer protocol
type TCPPeer struct {
	conn net.Conn

	// if we dial and connection is established then outBound is true
	// if we accept a connection then outBound is false because it will be and inbound connection
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Printf("TCP accept error: %v\n", err)
		}

		go t.handleConnection(conn)
	}
}

func (t *TCPTransport) handleConnection(connection net.Conn) {
	log.Printf("New incoming connection %v+v\n", connection)
}
