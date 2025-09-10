package main

import (
	"fmt"
	"log"

	"github.com/harhvardha/peer-to-peer-distributed-storage/p2p"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	//fmt.Println("doing some logic with peer outside of TCPTransport")
	return nil
}

func main() {
	tcpOps := p2p.TCPTransportOps{
		ListenAddr:    ":3000",
		Decoder:       p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
		OnPeer:        OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOps)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
