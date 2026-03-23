// Package mta implements the GD92 Message Transfer Agent layer,
// providing framed envelope transport over TCP connections.
package mta

import "time"

// AckType determines which acknowledgement character the MTA uses.
type AckType byte

const (
	// AckMaster means this MTA transmits ACKM (0x06) and expects ACKS (0x07).
	AckMaster AckType = 'M'
	// AckSlave means this MTA transmits ACKS (0x07) and expects ACKM (0x06).
	AckSlave AckType = 'S'
)

// Config holds MTA parameters as defined in spec sections 5.6/C.8.
type Config struct {
	// FrameDuration is the maximum time to send/receive a complete frame.
	FrameDuration time.Duration
	// FrameTimeout is the maximum time to wait for a frame acknowledgement.
	FrameTimeout time.Duration
	// RetriesAllowed is the maximum number of retransmission attempts.
	RetriesAllowed int
	// VerifyPeriod is how often to send ENQ if no traffic received.
	VerifyPeriod time.Duration
	// VerifyTimeout is the maximum inactivity before declaring link failure.
	VerifyTimeout time.Duration
	// AckType determines whether this side is master (ACKM) or slave (ACKS).
	AckType AckType
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		FrameDuration:  10 * time.Second,
		FrameTimeout:   5 * time.Second,
		RetriesAllowed: 3,
		VerifyPeriod:   30 * time.Second,
		VerifyTimeout:  90 * time.Second,
		AckType:        AckMaster,
	}
}

// ackCharTx returns the acknowledgement byte this MTA transmits.
func (c Config) ackCharTx() byte {
	if c.AckType == AckSlave {
		return 0x07 // ACKS
	}
	return 0x06 // ACKM
}

// ackCharRx returns the acknowledgement byte this MTA expects to receive.
func (c Config) ackCharRx() byte {
	if c.AckType == AckSlave {
		return 0x06 // ACKM
	}
	return 0x07 // ACKS
}
