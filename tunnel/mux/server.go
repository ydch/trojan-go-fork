package mux

import (
	"context"

	"github.com/xtaci/smux"

	"github.com/Potterli20/trojan-go-fork/common"
	"github.com/Potterli20/trojan-go-fork/log"
	"github.com/Potterli20/trojan-go-fork/tunnel"
)

// Server is a smux server
type Server struct {
	underlay tunnel.Server
	connChan chan tunnel.Conn
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) acceptConnWorker() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})
		if err != nil {
			log.Debug(err)
			select {
			case <-s.ctx.Done():
				return
			default:
			}
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn tunnel.Conn) {
	smuxConfig := smux.DefaultConfig()
	// smuxConfig.KeepAliveDisabled = true
	smuxSession, err := smux.Server(conn, smuxConfig)
	if err != nil {
		log.Error(err)
		return
	}
	go s.handleSession(smuxSession, conn)
}

func (s *Server) handleSession(session *smux.Session, conn tunnel.Conn) {
	defer session.Close()
	defer conn.Close()
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			log.Error(err)
			return
		}
		select {
		case s.connChan <- &Conn{
			rwc:  stream,
			Conn: conn,
		}:
		case <-s.ctx.Done():
			log.Debug("exiting")
			return
		}
	}
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("mux server closed")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay: underlay,
		ctx:      ctx,
		cancel:   cancel,
		connChan: make(chan tunnel.Conn, 32),
	}
	go server.acceptConnWorker()
	log.Debug("mux server created")
	return server, nil
}
