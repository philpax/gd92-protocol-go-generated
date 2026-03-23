package gd92

// Frame protocol constants as defined in spec section 5.6.4.
const (
	SOH  = 0x01 // Start of Header - marks beginning of envelope frame
	EOT  = 0x04 // End of Transmission - marks end of envelope frame
	ENQ  = 0x05 // Enquiry - link verification
	ACKM = 0x06 // Acknowledgement (Master)
	ACKS = 0x07 // Acknowledgement (Slave)
)

// WrapFrame wraps an envelope in SOH/EOT framing: <SOH><envelope><EOT>.
func WrapFrame(envelope []byte) []byte {
	frame := make([]byte, len(envelope)+2)
	frame[0] = SOH
	copy(frame[1:], envelope)
	frame[len(frame)-1] = EOT
	return frame
}

// UnwrapFrame strips SOH/EOT framing and returns the envelope bytes.
// Returns nil if the data is not a valid frame.
func UnwrapFrame(frame []byte) []byte {
	if len(frame) < 3 {
		return nil
	}
	if frame[0] != SOH || frame[len(frame)-1] != EOT {
		return nil
	}
	out := make([]byte, len(frame)-2)
	copy(out, frame[1:len(frame)-1])
	return out
}
