package gd92

import "testing"

func TestAddressRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		addr CommsAddress
	}{
		{"zero", CommsAddress{0, 0, 0}},
		{"max", CommsAddress{255, 1023, 63}},
		{"typical", CommsAddress{5, 42, 3}},
		{"port_only", CommsAddress{0, 0, 63}},
		{"node_boundary", CommsAddress{1, 4, 0}}, // node=4 uses second byte bits
		{"node_hi_lo", CommsAddress{10, 513, 33}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.addr.MarshalAddress()
			got := UnmarshalAddress(b)
			if !got.Equal(tt.addr) {
				t.Fatalf("round-trip failed: %v -> %v -> %v", tt.addr, b, got)
			}
		})
	}
}

func TestAddressBitPacking(t *testing.T) {
	// Brigade=5, Node=42 (0b00_1010_10), Port=3 (0b00_0011)
	// byte[0] = 5
	// byte[1] = 42>>2 = 10
	// byte[2] = (42&3)<<6 | 3 = 2<<6 | 3 = 128+3 = 131
	a := CommsAddress{Brigade: 5, Node: 42, Port: 3}
	b := a.MarshalAddress()
	if b[0] != 5 {
		t.Fatalf("byte[0]: expected 5, got %d", b[0])
	}
	if b[1] != 10 {
		t.Fatalf("byte[1]: expected 10, got %d", b[1])
	}
	if b[2] != 131 {
		t.Fatalf("byte[2]: expected 131, got %d", b[2])
	}
}

func TestAddressString(t *testing.T) {
	a := CommsAddress{Brigade: 5, Node: 42, Port: 3}
	s := a.String()
	if s != "B5/N42/P3" {
		t.Fatalf("expected B5/N42/P3, got %s", s)
	}
}

func TestAddressDecoderEncoder(t *testing.T) {
	orig := CommsAddress{Brigade: 100, Node: 500, Port: 30}

	enc := NewEncoder()
	enc.WriteCommsAddress(orig)

	dec := NewDecoder(enc.Bytes())
	got, err := dec.ReadCommsAddress()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(orig) {
		t.Fatalf("expected %v, got %v", orig, got)
	}
}
