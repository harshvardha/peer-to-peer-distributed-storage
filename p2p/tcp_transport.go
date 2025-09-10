package p2p

import (
	"fmt"
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

// TCPPeer implements Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOps struct {
	ListenAddr    string
	HandshakeFunc HandeshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOps
	listener net.Listener
	rpcch    chan RPC

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(ops TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOps: ops,
		rpcch:           make(chan RPC),
	}
}

// Consume implements the transport interface, which will return read-only channel
// for reading the incoming messages received from another peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.TCPTransportOps.ListenAddr)
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

		log.Printf("New incoming connection %v+v\n", conn)
		go t.handleConnection(conn)
	}
}

func (t *TCPTransport) handleConnection(connection net.Conn) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		connection.Close()
	}()

	peer := NewTCPPeer(connection, true)

	if err = t.TCPTransportOps.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// read loop
	rpc := RPC{}
	for {
		if err = t.Decoder.Decode(connection, &rpc); err != nil {
			return
		}

		rpc.From = connection.RemoteAddr().String()
		t.rpcch <- rpc
	}
}
