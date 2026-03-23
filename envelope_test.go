package gd92

import (
	"testing"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	env := &Envelope{
		Source:       CommsAddress{Brigade: 5, Node: 42, Port: 3},
		Destinations: []CommsAddress{{Brigade: 10, Node: 100, Port: 4}},
		Priority:     1,
		ProtVers:     1,
		AckReq:       true,
		Seq:          12345,
		MessageType:  2,
		Message:      []byte("HELLO"),
	}

	data, err := env.MarshalEnvelope()
	if err != nil {
		t.Fatal(err)
	}

	got, err := UnmarshalEnvelope(data)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Source.Equal(env.Source) {
		t.Fatalf("source: expected %v, got %v", env.Source, got.Source)
	}
	if len(got.Destinations) != 1 {
		t.Fatalf("destinations: expected 1, got %d", len(got.Destinations))
	}
	if !got.Destinations[0].Equal(env.Destinations[0]) {
		t.Fatalf("dest[0]: expected %v, got %v", env.Destinations[0], got.Destinations[0])
	}
	if got.Priority != env.Priority {
		t.Fatalf("priority: expected %d, got %d", env.Priority, got.Priority)
	}
	if got.ProtVers != env.ProtVers {
		t.Fatalf("prot_vers: expected %d, got %d", env.ProtVers, got.ProtVers)
	}
	if got.AckReq != env.AckReq {
		t.Fatalf("ack_req: expected %v, got %v", env.AckReq, got.AckReq)
	}
	if got.Seq != env.Seq {
		t.Fatalf("seq: expected %d, got %d", env.Seq, got.Seq)
	}
	if got.MessageType != env.MessageType {
		t.Fatalf("message_type: expected %d, got %d", env.MessageType, got.MessageType)
	}
	if string(got.Message) != string(env.Message) {
		t.Fatalf("message: expected %q, got %q", env.Message, got.Message)
	}
}

func TestEnvelopeMultipleDestinations(t *testing.T) {
	env := &Envelope{
		Source: CommsAddress{Brigade: 1, Node: 1, Port: 1},
		Destinations: []CommsAddress{
			{Brigade: 2, Node: 2, Port: 2},
			{Brigade: 3, Node: 3, Port: 3},
			{Brigade: 4, Node: 4, Port: 4},
		},
		Priority:    3,
		ProtVers:    1,
		AckReq:      false,
		Seq:         100,
		MessageType: 50,
		Message:     nil,
	}

	data, err := env.MarshalEnvelope()
	if err != nil {
		t.Fatal(err)
	}

	got, err := UnmarshalEnvelope(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Destinations) != 3 {
		t.Fatalf("expected 3 destinations, got %d", len(got.Destinations))
	}
	for i, dst := range got.Destinations {
		if !dst.Equal(env.Destinations[i]) {
			t.Fatalf("dest[%d]: expected %v, got %v", i, env.Destinations[i], dst)
		}
	}
	if got.AckReq {
		t.Fatal("expected ack_req=false")
	}
}

func TestEnvelopeBCC(t *testing.T) {
	env := &Envelope{
		Source:       CommsAddress{Brigade: 1, Node: 0, Port: 0},
		Destinations: []CommsAddress{{Brigade: 2, Node: 0, Port: 0}},
		Priority:     1,
		ProtVers:     1,
		Seq:          1,
		MessageType:  50,
		Message:      nil,
	}

	data, err := env.MarshalEnvelope()
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt a byte
	data[2] ^= 0xFF
	_, err = UnmarshalEnvelope(data)
	if err == nil {
		t.Fatal("expected BCC error on corrupted data")
	}
}

func TestEnvelopeEmptyMessage(t *testing.T) {
	env := &Envelope{
		Source:       CommsAddress{Brigade: 1, Node: 0, Port: 0},
		Destinations: []CommsAddress{{Brigade: 2, Node: 0, Port: 0}},
		Priority:     3,
		ProtVers:     1,
		Seq:          0,
		MessageType:  50, // ACK has empty body
		Message:      nil,
	}

	data, err := env.MarshalEnvelope()
	if err != nil {
		t.Fatal(err)
	}

	got, err := UnmarshalEnvelope(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Message) != 0 {
		t.Fatalf("expected empty message, got %d bytes", len(got.Message))
	}
}

func TestEnvelopeSize(t *testing.T) {
	env := &Envelope{
		Source:       CommsAddress{Brigade: 1, Node: 0, Port: 0},
		Destinations: []CommsAddress{{Brigade: 2, Node: 0, Port: 0}},
		Priority:     1,
		ProtVers:     1,
		Seq:          1,
		MessageType:  2,
		Message:      []byte("TEST"),
	}

	data, err := env.MarshalEnvelope()
	if err != nil {
		t.Fatal(err)
	}

	var header [5]byte
	copy(header[:], data[:5])
	size := EnvelopeSize(header)
	if size != len(data) {
		t.Fatalf("EnvelopeSize: expected %d, got %d", len(data), size)
	}
}

func TestEnvelopeInvalidDestCount(t *testing.T) {
	env := &Envelope{
		Source:       CommsAddress{},
		Destinations: nil,
		MessageType:  50,
	}
	_, err := env.MarshalEnvelope()
	if err == nil {
		t.Fatal("expected error for 0 destinations")
	}
}
