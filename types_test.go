package gd92

import (
	"testing"
)

func TestDecoderWord8(t *testing.T) {
	d := NewDecoder([]byte{0x42, 0xFF})
	v, err := d.ReadWord8()
	if err != nil {
		t.Fatal(err)
	}
	if v != 0x42 {
		t.Fatalf("expected 0x42, got 0x%02x", v)
	}
	v, err = d.ReadWord8()
	if err != nil {
		t.Fatal(err)
	}
	if v != 0xFF {
		t.Fatalf("expected 0xFF, got 0x%02x", v)
	}
	_, err = d.ReadWord8()
	if err != ErrShortRead {
		t.Fatalf("expected ErrShortRead, got %v", err)
	}
}

func TestDecoderWord16(t *testing.T) {
	d := NewDecoder([]byte{0x01, 0x02})
	v, err := d.ReadWord16()
	if err != nil {
		t.Fatal(err)
	}
	if v != 0x0102 {
		t.Fatalf("expected 0x0102, got 0x%04x", v)
	}
}

func TestDecoderWord32(t *testing.T) {
	d := NewDecoder([]byte{0x01, 0x02, 0x03, 0x04})
	v, err := d.ReadWord32()
	if err != nil {
		t.Fatal(err)
	}
	if v != 0x01020304 {
		t.Fatalf("expected 0x01020304, got 0x%08x", v)
	}
}

func TestDecoderBool(t *testing.T) {
	d := NewDecoder([]byte{0x00, 0x01, 0x02})
	v, err := d.ReadBool()
	if err != nil {
		t.Fatal(err)
	}
	if v != false {
		t.Fatal("expected false")
	}
	v, err = d.ReadBool()
	if err != nil {
		t.Fatal(err)
	}
	if v != true {
		t.Fatal("expected true")
	}
	_, err = d.ReadBool()
	if err == nil {
		t.Fatal("expected error for invalid bool value 2")
	}
}

func TestDecoderString(t *testing.T) {
	data := []byte{0x05, 'H', 'E', 'L', 'L', 'O'}
	d := NewDecoder(data)
	s, err := d.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if s != "HELLO" {
		t.Fatalf("expected HELLO, got %q", s)
	}
}

func TestDecoderStringEmpty(t *testing.T) {
	data := []byte{0x00}
	d := NewDecoder(data)
	s, err := d.ReadString()
	if err != nil {
		t.Fatal(err)
	}
	if s != "" {
		t.Fatalf("expected empty string, got %q", s)
	}
}

func TestDecoderTimeAndDate(t *testing.T) {
	data := []byte("15MAR96143022")
	d := NewDecoder(data)
	s, err := d.ReadTimeAndDate()
	if err != nil {
		t.Fatal(err)
	}
	if s != "15MAR96143022" {
		t.Fatalf("expected 15MAR96143022, got %q", s)
	}
}

func TestEncoderRoundTrip(t *testing.T) {
	enc := NewEncoder()
	enc.WriteWord8(0x42)
	enc.WriteWord16(0x1234)
	enc.WriteWord32(0xDEADBEEF)
	enc.WriteBool(true)
	enc.WriteBool(false)
	enc.WriteString("TEST")
	enc.WriteTimeAndDate("01JAN00120000")

	d := NewDecoder(enc.Bytes())

	w8, _ := d.ReadWord8()
	if w8 != 0x42 {
		t.Fatalf("word8: expected 0x42, got 0x%02x", w8)
	}

	w16, _ := d.ReadWord16()
	if w16 != 0x1234 {
		t.Fatalf("word16: expected 0x1234, got 0x%04x", w16)
	}

	w32, _ := d.ReadWord32()
	if w32 != 0xDEADBEEF {
		t.Fatalf("word32: expected 0xDEADBEEF, got 0x%08x", w32)
	}

	b1, _ := d.ReadBool()
	if !b1 {
		t.Fatal("expected true")
	}

	b2, _ := d.ReadBool()
	if b2 {
		t.Fatal("expected false")
	}

	s, _ := d.ReadString()
	if s != "TEST" {
		t.Fatalf("expected TEST, got %q", s)
	}

	td, _ := d.ReadTimeAndDate()
	if td != "01JAN00120000" {
		t.Fatalf("expected 01JAN00120000, got %q", td)
	}

	if d.Remaining() != 0 {
		t.Fatalf("expected 0 remaining, got %d", d.Remaining())
	}
}

func TestFixedASCIIRoundTrip(t *testing.T) {
	enc := NewEncoder()
	enc.WriteFixedASCII("HP", 3)

	d := NewDecoder(enc.Bytes())
	s, err := d.ReadFixedASCII(3)
	if err != nil {
		t.Fatal(err)
	}
	if s != "HP " {
		t.Fatalf("expected %q, got %q", "HP ", s)
	}
}
