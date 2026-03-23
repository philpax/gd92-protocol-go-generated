package gd92

import (
	"bytes"
	"testing"
)

func TestCompressSpecExample(t *testing.T) {
	// Spec A.3.3: "     HIGH STREET" (5 spaces + "HIGH STREET")
	input := []byte("     HIGH STREET")
	got := Compress(input)
	// Expected: ESC ' ' 0x05 H I G H ' ' S T R E E T
	expected := []byte{escByte, ' ', 0x05, 'H', 'I', 'G', 'H', ' ', 'S', 'T', 'R', 'E', 'E', 'T'}
	if !bytes.Equal(got, expected) {
		t.Fatalf("compress mismatch:\n  got:      %v\n  expected: %v", got, expected)
	}
}

func TestDecompressSpecExample(t *testing.T) {
	compressed := []byte{escByte, ' ', 0x05, 'H', 'I', 'G', 'H', ' ', 'S', 'T', 'R', 'E', 'E', 'T'}
	got, err := Decompress(compressed)
	if err != nil {
		t.Fatal(err)
	}
	expected := "     HIGH STREET"
	if got != expected {
		t.Fatalf("decompress: expected %q, got %q", expected, got)
	}
}

func TestCompressDecompressRoundTrip(t *testing.T) {
	tests := []string{
		"",
		"HELLO",
		"     HIGH STREET",
		"AAAAAAAAAA",               // 10 A's
		"ABCABC",                   // no runs >3
		"AAAAABBBB",                // two runs
		string([]byte{escByte}),    // literal ESC
		"test" + string([]byte{escByte}) + "end",
		"   ",  // run of exactly 3
		"    ", // run of exactly 4
	}
	for _, tc := range tests {
		compressed := Compress([]byte(tc))
		decompressed, err := Decompress(compressed)
		if err != nil {
			t.Fatalf("decompress error for %q: %v", tc, err)
		}
		if decompressed != tc {
			t.Fatalf("round-trip failed for %q: got %q", tc, decompressed)
		}
	}
}

func TestCompressLiteralESC(t *testing.T) {
	input := []byte{escByte}
	got := Compress(input)
	expected := []byte{escByte, escByte, 0x01}
	if !bytes.Equal(got, expected) {
		t.Fatalf("ESC literal: expected %v, got %v", expected, got)
	}
}

func TestCompressNoRunShortSequences(t *testing.T) {
	// Runs of 3 or fewer should not be compressed
	input := []byte("AAA")
	got := Compress(input)
	if !bytes.Equal(got, input) {
		t.Fatalf("3-char run should not compress: expected %v, got %v", input, got)
	}
}

func TestDecompressMalformed(t *testing.T) {
	// ESC followed by insufficient data
	_, err := Decompress([]byte{escByte, 'A'})
	if err == nil {
		t.Fatal("expected error for truncated ESC sequence")
	}
}
