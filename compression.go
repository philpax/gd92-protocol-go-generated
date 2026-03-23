package gd92

import (
	"bytes"
	"errors"
)

const escByte = 0x1B // ASCII ESC character

// ErrBadCompression is returned when compressed data is malformed.
var ErrBadCompression = errors.New("gd92: malformed compressed data")

// Compress applies GD92 ESC-based run-length encoding (spec A.3).
// Runs of more than 3 identical characters are encoded as ESC char count.
// Literal ESC characters are encoded as ESC ESC 0x01.
func Compress(data []byte) []byte {
	var out bytes.Buffer
	i := 0
	for i < len(data) {
		ch := data[i]

		// Handle literal ESC character
		if ch == escByte {
			out.WriteByte(escByte)
			out.WriteByte(escByte)
			out.WriteByte(0x01)
			i++
			continue
		}

		// Count run length
		runLen := 1
		for i+runLen < len(data) && data[i+runLen] == ch && runLen < 255 {
			runLen++
		}

		if runLen > 3 {
			out.WriteByte(escByte)
			out.WriteByte(ch)
			out.WriteByte(byte(runLen))
			i += runLen
		} else {
			// Output characters individually
			for j := 0; j < runLen; j++ {
				out.WriteByte(ch)
			}
			i += runLen
		}
	}
	return out.Bytes()
}

// Decompress reverses GD92 ESC-based run-length encoding (spec A.3).
// ESC ESC -> emit one ESC (consuming following 0x01)
// ESC char count -> emit char repeated count times
func Decompress(data []byte) (string, error) {
	var out bytes.Buffer
	i := 0
	for i < len(data) {
		if data[i] == escByte {
			if i+2 >= len(data) {
				return "", ErrBadCompression
			}
			ch := data[i+1]
			if ch == escByte {
				// Literal ESC: ESC ESC 0x01
				out.WriteByte(escByte)
				i += 3 // skip ESC ESC 0x01
			} else {
				// Run: ESC char count
				count := data[i+2]
				for j := 0; j < int(count); j++ {
					out.WriteByte(ch)
				}
				i += 3
			}
		} else {
			out.WriteByte(data[i])
			i++
		}
	}
	return out.String(), nil
}
