package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/harhvardha/peer-to-peer-distributed-storage/p2p"
)

type FileServerOps struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOps

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitCh chan struct{}
}

func NewFileServer(ops FileServerOps) *FileServer {
	storeOps := StoreOps{
		Root:              ops.StorageRoot,
		PathTransformFunc: ops.PathTransformFunc,
	}

	return &FileServer{
		FileServerOps: ops,
		store:         NewStore(storeOps),
		quitCh:        make(chan struct{}),
		peers:         make(map[string]p2p.Peer),
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	if err := s.store.Write(msg.Key, io.LimitReader(peer, int64(msg.Size))); err != nil {
		return err
	}

	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}

func (fs *FileServer) bootstrapNetwork() error {
	for _, addr := range fs.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			if err := fs.Transport.Dial(addr); err != nil {
				log.Println("Dial error: ", err)
			}
		}(addr)
	}

	return nil
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.bootstrapNetwork()
	fs.loop()
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	multiWriter := io.MultiWriter(peers...)
	return gob.NewEncoder(multiWriter).Encode(msg)
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: 15,
		},
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range fs.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(3 * time.Second)

	payload := []byte("THIS LARGE FILE")
	for _, peer := range fs.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil

	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// // storing the file on our local disk first
	// if err := fs.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// // then broadcasting to the available peers
	// msg := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// return fs.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: msg,
	// })
}

func (fs *FileServer) Stop() {
	close(fs.quitCh)
}

func (fs *FileServer) OnPeer(p p2p.Peer) error {
	fs.peerLock.Lock()
	defer fs.peerLock.Unlock()

	fs.peers[p.RemoteAddr().String()] = p
	log.Printf("Connected with peer with address: %s", p.RemoteAddr().String())
	return nil
}

func (fs *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit action")
		fs.Transport.Close()
	}()

	for {
		select {
		case rpc := <-fs.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
			}

			if err := fs.handleMessage(rpc.From, &msg); err != nil {
				log.Println(err)
				return
			}
			// fmt.Printf("%+v\n", msg.Payload)

			// peer, ok := fs.peers[rpc.From]
			// if !ok {
			// 	panic("peer not found in peer map")
			// }

			// b := make([]byte, 1000)
			// if _, err := peer.Read(b); err != nil {
			// 	panic(err)
			// }

			// fmt.Printf("%s\n", string(b))
			// peer.(*p2p.TCPPeer).Wg.Done()

			// if err := fs.handleMessage(&msg); err != nil {
			// 	log.Println(err)
			// }
		case <-fs.quitCh:
			return
		}
	}
}

func (fs *FileServer) Store(key string, r io.Reader) error {
	return fs.store.Write(key, r)
}
