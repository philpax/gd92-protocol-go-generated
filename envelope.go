package gd92

import (
	"errors"
	"fmt"
)

// Envelope errors.
var (
	ErrBadBCC       = errors.New("gd92: BCC verification failed")
	ErrBadDestCount = errors.New("gd92: dest_count must be 1-63")
	ErrMsgTooLong   = errors.New("gd92: message exceeds 1023 bytes")
)

// Envelope is the GD92 message envelope as defined in spec section 3.3.
// Wire layout:
//
//	source(3) | count&length(2) | destinations(N*3) | prot&priority(1) | ack&seq(2) | message_type(1) | message(L) | bcc(1)
type Envelope struct {
	Source       CommsAddress
	Destinations []CommsAddress // 1-63 destinations
	Priority     uint8          // 1-9 (1=highest)
	ProtVers     uint8          // 1-15
	AckReq       bool
	Seq          uint16 // 0-32767
	MessageType  uint8
	Message      []byte // 0-1023 bytes
}

// MarshalEnvelope encodes an envelope to wire format including BCC.
func (env *Envelope) MarshalEnvelope() ([]byte, error) {
	destCount := len(env.Destinations)
	if destCount < 1 || destCount > 63 {
		return nil, fmt.Errorf("%w: got %d", ErrBadDestCount, destCount)
	}
	msgLen := len(env.Message)
	if msgLen > 1023 {
		return nil, fmt.Errorf("%w: got %d", ErrMsgTooLong, msgLen)
	}

	enc := NewEncoder()

	// Source address (3 bytes)
	enc.WriteCommsAddress(env.Source)

	// count&length (word16): bits 0-9 = length, bits 10-15 = dest_count
	countLength := uint16(msgLen) | (uint16(destCount) << 10)
	enc.WriteWord16(countLength)

	// Destinations (N * 3 bytes)
	for _, dst := range env.Destinations {
		enc.WriteCommsAddress(dst)
	}

	// prot&priority (word8): bits 0-3 = priority, bits 4-7 = prot_vers
	protPriority := (env.ProtVers << 4) | (env.Priority & 0x0F)
	enc.WriteWord8(protPriority)

	// ack&seq (word16): bits 0-14 = seq, bit 15 = ack_req
	ackSeq := env.Seq & 0x7FFF
	if env.AckReq {
		ackSeq |= 0x8000
	}
	enc.WriteWord16(ackSeq)

	// message_type (word8)
	enc.WriteWord8(env.MessageType)

	// message (L bytes)
	enc.WriteBytes(env.Message)

	// BCC: XOR of all preceding bytes
	data := enc.Bytes()
	bcc := computeBCC(data)
	enc.WriteWord8(bcc)

	return enc.Bytes(), nil
}

// UnmarshalEnvelope decodes an envelope from wire format, verifying BCC.
func UnmarshalEnvelope(data []byte) (*Envelope, error) {
	if len(data) < 10 {
		// Minimum: source(3) + count&length(2) + dest(3) + prot_priority(1) + ack_seq(2) + msg_type(1) + bcc(1) = 13
		// But with 0-length message: 3+2+3+1+2+1+0+1 = 13
		return nil, ErrShortRead
	}

	// Verify BCC
	bcc := computeBCC(data[:len(data)-1])
	if bcc != data[len(data)-1] {
		return nil, fmt.Errorf("%w: computed 0x%02x, got 0x%02x", ErrBadBCC, bcc, data[len(data)-1])
	}

	d := NewDecoder(data[:len(data)-1]) // exclude BCC

	env := &Envelope{}

	// Source
	var err error
	env.Source, err = d.ReadCommsAddress()
	if err != nil {
		return nil, err
	}

	// count&length
	cl, err := d.ReadWord16()
	if err != nil {
		return nil, err
	}
	msgLen := int(cl & 0x03FF)
	destCount := int(cl >> 10)
	if destCount < 1 || destCount > 63 {
		return nil, fmt.Errorf("%w: got %d", ErrBadDestCount, destCount)
	}

	// Destinations
	env.Destinations = make([]CommsAddress, destCount)
	for i := 0; i < destCount; i++ {
		env.Destinations[i], err = d.ReadCommsAddress()
		if err != nil {
			return nil, err
		}
	}

	// prot&priority
	pp, err := d.ReadWord8()
	if err != nil {
		return nil, err
	}
	env.Priority = pp & 0x0F
	env.ProtVers = pp >> 4

	// ack&seq
	as, err := d.ReadWord16()
	if err != nil {
		return nil, err
	}
	env.Seq = as & 0x7FFF
	env.AckReq = (as & 0x8000) != 0

	// message_type
	env.MessageType, err = d.ReadWord8()
	if err != nil {
		return nil, err
	}

	// message
	if d.Remaining() < msgLen {
		return nil, ErrShortRead
	}
	env.Message, err = d.ReadBytes(msgLen)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// EnvelopeSize calculates the total envelope size (including BCC) from the
// first 5 bytes of the envelope (source + count&length). This is used by
// the MTA layer to know how many bytes to read after SOH.
func EnvelopeSize(header [5]byte) int {
	cl := uint16(header[3])<<8 | uint16(header[4])
	msgLen := int(cl & 0x03FF)
	destCount := int(cl >> 10)
	// source(3) + count&length(2) + destinations(N*3) + prot_priority(1) + ack_seq(2) + msg_type(1) + message(L) + bcc(1)
	return 3 + 2 + destCount*3 + 1 + 2 + 1 + msgLen + 1
}

// computeBCC calculates the block check character by XORing all bytes.
func computeBCC(data []byte) byte {
	var bcc byte
	for _, b := range data {
		bcc ^= b
	}
	return bcc
}
