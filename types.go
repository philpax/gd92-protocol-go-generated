package gd92

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Errors returned by Decoder when there is insufficient data.
var (
	ErrShortRead  = errors.New("gd92: not enough data")
	ErrBadString  = errors.New("gd92: string length exceeds remaining data")
	ErrBadBool    = errors.New("gd92: boolean value not 0 or 1")
)

// Decoder reads GD92 binary data types from a byte slice.
// All multi-byte integers are big-endian unsigned as per the spec.
type Decoder struct {
	data []byte
	pos  int
}

// NewDecoder creates a Decoder over the given data.
func NewDecoder(data []byte) *Decoder {
	return &Decoder{data: data}
}

// Remaining returns the number of unread bytes.
func (d *Decoder) Remaining() int {
	return len(d.data) - d.pos
}

// Pos returns the current read position.
func (d *Decoder) Pos() int {
	return d.pos
}

// ReadWord8 reads a single unsigned byte.
func (d *Decoder) ReadWord8() (uint8, error) {
	if d.Remaining() < 1 {
		return 0, ErrShortRead
	}
	v := d.data[d.pos]
	d.pos++
	return v, nil
}

// ReadWord16 reads a big-endian unsigned 16-bit integer.
func (d *Decoder) ReadWord16() (uint16, error) {
	if d.Remaining() < 2 {
		return 0, ErrShortRead
	}
	v := binary.BigEndian.Uint16(d.data[d.pos:])
	d.pos += 2
	return v, nil
}

// ReadWord32 reads a big-endian unsigned 32-bit integer.
func (d *Decoder) ReadWord32() (uint32, error) {
	if d.Remaining() < 4 {
		return 0, ErrShortRead
	}
	v := binary.BigEndian.Uint32(d.data[d.pos:])
	d.pos += 4
	return v, nil
}

// ReadBool reads a GD92 boolean (0=false, 1=true).
func (d *Decoder) ReadBool() (bool, error) {
	b, err := d.ReadWord8()
	if err != nil {
		return false, err
	}
	switch b {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("%w: got %d", ErrBadBool, b)
	}
}

// ReadBytes reads exactly n raw bytes.
func (d *Decoder) ReadBytes(n int) ([]byte, error) {
	if d.Remaining() < n {
		return nil, ErrShortRead
	}
	out := make([]byte, n)
	copy(out, d.data[d.pos:d.pos+n])
	d.pos += n
	return out, nil
}

// ReadString reads a <string>: 1-byte count prefix followed by that many ASCII bytes.
func (d *Decoder) ReadString() (string, error) {
	count, err := d.ReadWord8()
	if err != nil {
		return "", err
	}
	if d.Remaining() < int(count) {
		return "", ErrBadString
	}
	s := string(d.data[d.pos : d.pos+int(count)])
	d.pos += int(count)
	return s, nil
}

// ReadCompressedString reads a <compressed_string>: 1-byte count prefix
// followed by that many bytes of ESC-compressed data, then decompresses.
func (d *Decoder) ReadCompressedString() (string, error) {
	count, err := d.ReadWord8()
	if err != nil {
		return "", err
	}
	if d.Remaining() < int(count) {
		return "", ErrBadString
	}
	compressed := d.data[d.pos : d.pos+int(count)]
	d.pos += int(count)
	return Decompress(compressed)
}

// ReadLongCompString reads a <long_comp_string>: 2-byte count prefix
// followed by that many bytes of ESC-compressed data, then decompresses.
func (d *Decoder) ReadLongCompString() (string, error) {
	count, err := d.ReadWord16()
	if err != nil {
		return "", err
	}
	if d.Remaining() < int(count) {
		return "", ErrBadString
	}
	compressed := d.data[d.pos : d.pos+int(count)]
	d.pos += int(count)
	return Decompress(compressed)
}

// ReadTimeAndDate reads a 13-byte fixed ASCII time_and_date field
// in the format "DDMMMYYHHMMSS" (no count prefix).
func (d *Decoder) ReadTimeAndDate() (string, error) {
	if d.Remaining() < 13 {
		return "", ErrShortRead
	}
	s := string(d.data[d.pos : d.pos+13])
	d.pos += 13
	return s, nil
}

// ReadFixedASCII reads exactly n bytes as a fixed-length ASCII field.
func (d *Decoder) ReadFixedASCII(n int) (string, error) {
	b, err := d.ReadBytes(n)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Encoder builds GD92 binary data.
type Encoder struct {
	buf []byte
}

// NewEncoder creates a new Encoder.
func NewEncoder() *Encoder {
	return &Encoder{}
}

// Bytes returns the encoded data.
func (e *Encoder) Bytes() []byte {
	return e.buf
}

// Len returns the current length of encoded data.
func (e *Encoder) Len() int {
	return len(e.buf)
}

// WriteWord8 appends a single byte.
func (e *Encoder) WriteWord8(v uint8) {
	e.buf = append(e.buf, v)
}

// WriteWord16 appends a big-endian 16-bit integer.
func (e *Encoder) WriteWord16(v uint16) {
	e.buf = append(e.buf, byte(v>>8), byte(v))
}

// WriteWord32 appends a big-endian 32-bit integer.
func (e *Encoder) WriteWord32(v uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	e.buf = append(e.buf, b...)
}

// WriteBool appends a GD92 boolean (0 or 1).
func (e *Encoder) WriteBool(v bool) {
	if v {
		e.buf = append(e.buf, 1)
	} else {
		e.buf = append(e.buf, 0)
	}
}

// WriteBytes appends raw bytes.
func (e *Encoder) WriteBytes(b []byte) {
	e.buf = append(e.buf, b...)
}

// WriteString writes a <string>: 1-byte count prefix followed by the ASCII data.
func (e *Encoder) WriteString(s string) {
	e.buf = append(e.buf, byte(len(s)))
	e.buf = append(e.buf, []byte(s)...)
}

// WriteCompressedString writes a <compressed_string>: compress the string,
// then write 1-byte count prefix followed by the compressed data.
func (e *Encoder) WriteCompressedString(s string) {
	compressed := Compress([]byte(s))
	e.buf = append(e.buf, byte(len(compressed)))
	e.buf = append(e.buf, compressed...)
}

// WriteLongCompString writes a <long_comp_string>: compress the string,
// then write 2-byte count prefix followed by the compressed data.
func (e *Encoder) WriteLongCompString(s string) {
	compressed := Compress([]byte(s))
	e.WriteWord16(uint16(len(compressed)))
	e.buf = append(e.buf, compressed...)
}

// WriteTimeAndDate writes a 13-byte fixed ASCII time_and_date field.
// The string must be exactly 13 bytes in "DDMMMYYHHMMSS" format.
func (e *Encoder) WriteTimeAndDate(s string) {
	padded := s
	if len(padded) < 13 {
		padded = padded + strings.Repeat(" ", 13-len(padded))
	}
	e.buf = append(e.buf, []byte(padded[:13])...)
}

// WriteFixedASCII writes exactly n bytes, right-padding with spaces if needed.
func (e *Encoder) WriteFixedASCII(s string, n int) {
	padded := s
	if len(padded) < n {
		padded = padded + strings.Repeat(" ", n-len(padded))
	}
	e.buf = append(e.buf, []byte(padded[:n])...)
}
