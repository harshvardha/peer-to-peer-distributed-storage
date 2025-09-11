package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// It represents the remote node using TCP as transport layer protocol
type TCPPeer struct {
	// underlying connection of peer which in this case is TCP
	net.Conn

	// if we dial and connection is established then outBound is true
	// if we accept a connection then outBound is false because it will be and inbound connection
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
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

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConnection(conn, true)
	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.TCPTransportOps.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
	log.Printf("TCP transport listening on port: %s", t.ListenAddr)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			log.Printf("TCP accept error: %v\n", err)
		}

		go t.handleConnection(conn, false)
	}
}

func (t *TCPTransport) handleConnection(connection net.Conn, outbound bool) {
	var err error
	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		connection.Close()
	}()

	peer := NewTCPPeer(connection, outbound)

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
		peer.Wg.Add(1)
		fmt.Println("waiting till stream is done")
		t.rpcch <- rpc
		peer.Wg.Wait()
		fmt.Println("stream done continuing normal read loop")
	}
}
