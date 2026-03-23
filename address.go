package gd92

import "fmt"

// CommsAddress represents a GD92 communications address.
// It is encoded as 3 bytes (24 bits):
//
//	bits  0-7:  Brigade identifier (0-255)
//	bits  8-17: Node identifier (0-1023)
//	bits 18-23: Port identifier (0-63)
type CommsAddress struct {
	Brigade uint8
	Node    uint16 // 0-1023
	Port    uint8  // 0-63
}

// MarshalAddress encodes the address into 3 bytes.
func (a CommsAddress) MarshalAddress() [3]byte {
	var b [3]byte
	b[0] = a.Brigade
	b[1] = byte(a.Node >> 2)
	b[2] = byte((a.Node&0x03)<<6) | (a.Port & 0x3F)
	return b
}

// UnmarshalAddress decodes a 3-byte address.
func UnmarshalAddress(b [3]byte) CommsAddress {
	return CommsAddress{
		Brigade: b[0],
		Node:    uint16(b[1])<<2 | uint16(b[2]>>6),
		Port:    b[2] & 0x3F,
	}
}

// String returns a human-readable representation like "B5/N42/P3".
func (a CommsAddress) String() string {
	return fmt.Sprintf("B%d/N%d/P%d", a.Brigade, a.Node, a.Port)
}

// Equal reports whether two addresses are identical.
func (a CommsAddress) Equal(other CommsAddress) bool {
	return a.Brigade == other.Brigade && a.Node == other.Node && a.Port == other.Port
}

// ReadCommsAddress reads a 3-byte comms address from the decoder.
func (d *Decoder) ReadCommsAddress() (CommsAddress, error) {
	b, err := d.ReadBytes(3)
	if err != nil {
		return CommsAddress{}, err
	}
	return UnmarshalAddress([3]byte{b[0], b[1], b[2]}), nil
}

// WriteCommsAddress writes a 3-byte comms address.
func (e *Encoder) WriteCommsAddress(a CommsAddress) {
	b := a.MarshalAddress()
	e.buf = append(e.buf, b[0], b[1], b[2])
}
