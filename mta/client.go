package mta

import (
	"context"
	"gd92"
	"log"
	"net"
)

// Client connects to a remote GD92 MTA and provides send/receive capabilities.
type Client struct {
	conn     *Conn
	received chan *ReceivedEnvelope
	done     chan struct{}
}

// Dial connects to a remote GD92 MTA at the given address.
func Dial(ctx context.Context, addr string, config Config) (*Client, error) {
	var d net.Dialer
	raw, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	conn := NewConn(raw, config)
	cl := &Client{
		conn:     conn,
		received: make(chan *ReceivedEnvelope, 64),
		done:     make(chan struct{}),
	}

	go cl.readLoop(ctx)
	return cl, nil
}

// Send transmits an envelope to the remote MTA.
func (cl *Client) Send(ctx context.Context, env *gd92.Envelope) error {
	return cl.conn.SendEnvelope(ctx, env)
}

// Received returns a channel that delivers envelopes received from the remote MTA.
func (cl *Client) Received() <-chan *ReceivedEnvelope {
	return cl.received
}

// Conn returns the underlying MTA connection.
func (cl *Client) Conn() *Conn {
	return cl.conn
}

// Close closes the connection.
func (cl *Client) Close() error {
	err := cl.conn.Close()
	<-cl.done
	return err
}

// readLoop continuously reads envelopes and sends them to the received channel.
func (cl *Client) readLoop(ctx context.Context) {
	defer close(cl.done)
	defer close(cl.received)

	for {
		env, err := cl.conn.ReadEnvelope(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			select {
			case cl.received <- &ReceivedEnvelope{Err: err, Conn: cl.conn}:
			case <-ctx.Done():
			}
			return
		}

		var msg gd92.Message
		parsed, parseErr := gd92.ParseEnvelopeMessage(env)
		if parseErr != nil {
			log.Printf("mta: parse message type %d: %v", env.MessageType, parseErr)
		} else {
			msg = parsed
		}

		select {
		case cl.received <- &ReceivedEnvelope{
			Envelope: env,
			Message:  msg,
			Conn:     cl.conn,
		}:
		case <-ctx.Done():
			return
		}
	}
}
