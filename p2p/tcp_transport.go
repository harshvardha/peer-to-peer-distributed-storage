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

type TCPTransportOps struct {
	ListenAddr    string
	HandshakeFunc HandeshakeFunc
	Decoder       Decoder
}

type TCPTransport struct {
	TCPTransportOps
	listener net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(ops TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOps: ops,
	}
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
	peer := NewTCPPeer(connection, true)

	if err := t.TCPTransportOps.HandshakeFunc(peer); err != nil {
		connection.Close()
		log.Printf("TCP handshake error: %s\n", err)
		return
	}

	// read loop
	msg := &RPC{}
	for {
		if err := t.Decoder.Decode(connection, msg); err != nil {
			log.Println("Error reading from connection: %v", err)
			continue
		}

		msg.From = connection.RemoteAddr()
		fmt.Printf("message: %+v\n", msg)
	}
}
