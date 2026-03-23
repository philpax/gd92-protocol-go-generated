package mta

import (
	"context"
	"gd92"
	"log"
	"net"
	"sync"
)

// ReceivedEnvelope bundles a received envelope with its parsed message
// and a reference to the connection for sending replies.
type ReceivedEnvelope struct {
	Envelope *gd92.Envelope
	Message  gd92.Message // parsed message, nil if parsing failed
	Conn     *Conn
	Err      error // non-nil if a connection-level error occurred
}

// Server listens for incoming GD92 connections and delivers received
// envelopes via a channel.
type Server struct {
	listener net.Listener
	config   Config
	received chan *ReceivedEnvelope

	mu    sync.Mutex
	conns []*Conn
}

// Listen creates and starts a Server on the given address.
func Listen(addr string, config Config) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: ln,
		config:   config,
		received: make(chan *ReceivedEnvelope, 64),
	}
	return s, nil
}

// Received returns a channel that delivers received envelopes from all connections.
func (s *Server) Received() <-chan *ReceivedEnvelope {
	return s.received
}

// Addr returns the listener's address.
func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

// Serve accepts connections and reads envelopes until the context is cancelled.
// This method blocks.
func (s *Server) Serve(ctx context.Context) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	defer close(s.received)

	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		raw, err := s.listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		conn := NewConn(raw, s.config)
		s.mu.Lock()
		s.conns = append(s.conns, conn)
		s.mu.Unlock()

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConn(ctx, conn)
		}()
	}
}

// handleConn reads envelopes from a connection and sends them to the received channel.
func (s *Server) handleConn(ctx context.Context, conn *Conn) {
	defer conn.Close()

	for {
		env, err := conn.ReadEnvelope(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			select {
			case s.received <- &ReceivedEnvelope{Err: err, Conn: conn}:
			case <-ctx.Done():
			}
			return
		}

		// Parse message
		var msg gd92.Message
		parsed, parseErr := gd92.ParseEnvelopeMessage(env)
		if parseErr != nil {
			log.Printf("mta: parse message type %d: %v", env.MessageType, parseErr)
		} else {
			msg = parsed
		}

		select {
		case s.received <- &ReceivedEnvelope{
			Envelope: env,
			Message:  msg,
			Conn:     conn,
		}:
		case <-ctx.Done():
			return
		}
	}
}

// Close closes the listener and all active connections.
func (s *Server) Close() error {
	err := s.listener.Close()
	s.mu.Lock()
	for _, c := range s.conns {
		c.Close()
	}
	s.mu.Unlock()
	return err
}
