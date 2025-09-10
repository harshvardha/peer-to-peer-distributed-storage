package main

import (
	"log"

	"github.com/harhvardha/peer-to-peer-distributed-storage/p2p"
)

func main() {
	tcpOps := p2p.TCPTransportOps{
		ListenAddr:    ":3000",
		Decoder:       p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
	}

	tr := p2p.NewTCPTransport(tcpOps)
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
