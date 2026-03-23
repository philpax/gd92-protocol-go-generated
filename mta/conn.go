package mta

import (
	"context"
	"errors"
	"fmt"
	"gd92"
	"net"
	"sync"
	"time"
)

// Conn wraps a net.Conn to provide GD92 frame-level transport.
// It handles SOH/EOT framing, ACK handshaking, ENQ link verification,
// and retry logic as defined in spec section 5.6.4.
//
// A single read goroutine demultiplexes incoming bytes: ack characters
// are routed to the ack channel, and SOH-framed envelopes are routed
// to the envelope channel.
type Conn struct {
	raw    net.Conn
	config Config

	writeMu sync.Mutex

	envelopes chan *gd92.Envelope // received envelopes
	acks      chan struct{}       // received ack characters
	readErr   chan error          // fatal read error

	lastRx   time.Time
	lastRxMu sync.Mutex

	stopVerify chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

// NewConn creates a new MTA connection wrapper and starts background goroutines.
func NewConn(raw net.Conn, config Config) *Conn {
	c := &Conn{
		raw:        raw,
		config:     config,
		lastRx:     time.Now(),
		stopVerify: make(chan struct{}),
		envelopes:  make(chan *gd92.Envelope, 16),
		acks:       make(chan struct{}, 16),
		readErr:    make(chan error, 1),
	}
	c.wg.Add(2)
	go c.readLoop()
	go c.verifyLoop()
	return c
}

// Close shuts down the connection and stops background goroutines.
func (c *Conn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		close(c.stopVerify)
		err = c.raw.Close()
		c.wg.Wait()
	})
	return err
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.raw.RemoteAddr()
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.raw.LocalAddr()
}

// SendEnvelope marshals and transmits an envelope with retry logic.
func (c *Conn) SendEnvelope(ctx context.Context, env *gd92.Envelope) error {
	data, err := env.MarshalEnvelope()
	if err != nil {
		return fmt.Errorf("mta: marshal envelope: %w", err)
	}
	frame := gd92.WrapFrame(data)

	for retry := 0; retry <= c.config.RetriesAllowed; retry++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := c.sendFrame(frame); err != nil {
			if retry < c.config.RetriesAllowed {
				continue
			}
			return fmt.Errorf("mta: send failed: %w", err)
		}

		// Wait for ack from the demux read loop
		select {
		case <-c.acks:
			return nil
		case err := <-c.readErr:
			return fmt.Errorf("mta: connection error while waiting for ack: %w", err)
		case <-time.After(c.config.FrameTimeout):
			if retry < c.config.RetriesAllowed {
				continue
			}
			return errors.New("mta: send failed: ack timeout after retries exhausted")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return errors.New("mta: send failed: retries exhausted")
}

// ReadEnvelope returns the next received envelope from the connection.
func (c *Conn) ReadEnvelope(ctx context.Context) (*gd92.Envelope, error) {
	select {
	case env := <-c.envelopes:
		return env, nil
	case err := <-c.readErr:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// readLoop is the single reader goroutine that demuxes the byte stream.
func (c *Conn) readLoop() {
	defer c.wg.Done()

	buf := make([]byte, 1)
	for {
		// Read one byte at a time to detect frame starts and ack chars
		c.raw.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := c.raw.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case c.readErr <- err:
			default:
			}
			return
		}
		if n == 0 {
			continue
		}

		b := buf[0]

		switch {
		case b == c.config.ackCharRx():
			// Ack from remote
			select {
			case c.acks <- struct{}{}:
			default:
			}

		case b == gd92.ENQ:
			// Link verification - respond and update timer
			c.sendAck()
			c.touchLastRx()

		case b == gd92.SOH:
			// Start of envelope frame
			env, err := c.readEnvelopeAfterSOH()
			if err != nil {
				// Invalid frame - discard, don't ack
				continue
			}
			c.sendAck()
			c.touchLastRx()
			select {
			case c.envelopes <- env:
			default:
				// Drop if channel full (shouldn't happen in practice)
			}

		default:
			// Unexpected byte, skip
		}
	}
}

// readEnvelopeAfterSOH reads an envelope after SOH has been consumed.
func (c *Conn) readEnvelopeAfterSOH() (*gd92.Envelope, error) {
	deadline := time.Now().Add(c.config.FrameDuration)
	c.raw.SetReadDeadline(deadline)

	// Read first 5 bytes to determine envelope size
	header := make([]byte, 5)
	if _, err := readFull(c.raw, header); err != nil {
		return nil, fmt.Errorf("mta: read header: %w", err)
	}

	var h [5]byte
	copy(h[:], header)
	totalSize := gd92.EnvelopeSize(h)

	remaining := totalSize - 5
	if remaining < 0 {
		return nil, errors.New("mta: invalid envelope size")
	}

	envData := make([]byte, totalSize)
	copy(envData, header)
	if remaining > 0 {
		if _, err := readFull(c.raw, envData[5:]); err != nil {
			return nil, fmt.Errorf("mta: read envelope body: %w", err)
		}
	}

	// Read EOT
	eot := make([]byte, 1)
	if _, err := readFull(c.raw, eot); err != nil {
		return nil, fmt.Errorf("mta: read EOT: %w", err)
	}
	if eot[0] != gd92.EOT {
		return nil, fmt.Errorf("mta: expected EOT (0x04), got 0x%02x", eot[0])
	}

	return gd92.UnmarshalEnvelope(envData)
}

// sendFrame writes a complete frame to the connection.
func (c *Conn) sendFrame(frame []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.raw.SetWriteDeadline(time.Now().Add(c.config.FrameDuration))
	_, err := c.raw.Write(frame)
	c.raw.SetWriteDeadline(time.Time{})
	return err
}

// sendAck sends the local ack character.
func (c *Conn) sendAck() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.raw.Write([]byte{c.config.ackCharTx()})
}

// touchLastRx updates the last received timestamp.
func (c *Conn) touchLastRx() {
	c.lastRxMu.Lock()
	c.lastRx = time.Now()
	c.lastRxMu.Unlock()
}

// verifyLoop sends ENQ when no traffic has been received for VerifyPeriod.
func (c *Conn) verifyLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.VerifyPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopVerify:
			return
		case <-ticker.C:
			c.lastRxMu.Lock()
			idle := time.Since(c.lastRx)
			c.lastRxMu.Unlock()

			if idle >= c.config.VerifyPeriod {
				c.sendENQ()
			}
		}
	}
}

// sendENQ sends an ENQ link verification byte.
func (c *Conn) sendENQ() {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.raw.Write([]byte{gd92.ENQ})
}

// readFull reads exactly len(buf) bytes from r.
func readFull(r net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
