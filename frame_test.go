package gd92

import (
	"bytes"
	"testing"
)

func TestWrapUnwrapFrame(t *testing.T) {
	envelope := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	frame := WrapFrame(envelope)

	if frame[0] != SOH {
		t.Fatalf("expected SOH at start, got 0x%02x", frame[0])
	}
	if frame[len(frame)-1] != EOT {
		t.Fatalf("expected EOT at end, got 0x%02x", frame[len(frame)-1])
	}

	unwrapped := UnwrapFrame(frame)
	if !bytes.Equal(unwrapped, envelope) {
		t.Fatalf("unwrap mismatch: expected %v, got %v", envelope, unwrapped)
	}
}

func TestUnwrapFrameInvalid(t *testing.T) {
	if UnwrapFrame([]byte{0x00, 0x01}) != nil {
		t.Fatal("expected nil for short frame")
	}
	if UnwrapFrame([]byte{0x00, 0x01, 0x04}) != nil {
		t.Fatal("expected nil for missing SOH")
	}
	if UnwrapFrame([]byte{0x01, 0x01, 0x00}) != nil {
		t.Fatal("expected nil for missing EOT")
	}
}

func TestFrameConstants(t *testing.T) {
	if SOH != 0x01 {
		t.Fatalf("SOH: expected 0x01, got 0x%02x", SOH)
	}
	if EOT != 0x04 {
		t.Fatalf("EOT: expected 0x04, got 0x%02x", EOT)
	}
	if ENQ != 0x05 {
		t.Fatalf("ENQ: expected 0x05, got 0x%02x", ENQ)
	}
	if ACKM != 0x06 {
		t.Fatalf("ACKM: expected 0x06, got 0x%02x", ACKM)
	}
	if ACKS != 0x07 {
		t.Fatalf("ACKS: expected 0x07, got 0x%02x", ACKS)
	}
}
