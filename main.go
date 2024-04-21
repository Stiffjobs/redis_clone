package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
)

const defaultListenAddress = ":5001"

type Config struct {
	ListenAddress string
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan []byte
}

func NewServer(cfg Config) *Server {
	if len(cfg.ListenAddress) == 0 {
		cfg.ListenAddress = defaultListenAddress

	}
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan []byte),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Config.ListenAddress)
	if err != nil {
		return err
	}

	s.ln = ln

	go s.Loop()

	slog.Info("server running", "listenAddr", s.ListenAddress)

	return s.AcceptLoop()

}

func (s *Server) handleRawMessage(rawMsg []byte) error {
	fmt.Println(string(rawMsg))
	return nil
}

func (s *Server) Loop() {
	for {
		select {
		case rawMsg := <-s.msgCh:
			if err := s.handleRawMessage(rawMsg); err != nil {
				slog.Error("error handle rawMsg", "err", err.Error())
			}
		case peer := <-s.addPeerCh:
			s.peers[peer] = true
		case <-s.quitCh:
			return
		}
	}
}

func (s *Server) AcceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	slog.Info("new peer connected", "remoteAddress", conn.RemoteAddr())

	if err := peer.readLoop(); err != nil {
		slog.Error("read loop err", "err", err.Error(), "remoteAddr", conn.RemoteAddr())
	}
}

func main() {
	server := NewServer(Config{})
	log.Fatal(server.Start())
}
