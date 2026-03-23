package gd92

import "fmt"

// Message type constants as defined in spec Appendix B, Table B-1.
const (
	MsgMobiliseCommand      = 1
	MsgMobiliseMessage      = 2
	MsgPageOfficer          = 3
	MsgAreaPageMessage      = 4
	MsgResourceStatusReq    = 5
	MsgMobiliseMessageAlt   = 6  // Hampshire/Kent format, undefined
	MsgActivatePeripheral   = 7
	MsgDeactivatePeripheral = 8
	MsgPeripheralStatusReq  = 9
	MsgResetRequest         = 10
	MsgResourceStatus       = 20
	MsgDutyStaffingUpdate   = 21
	MsgLogUpdate            = 22
	MsgStop                 = 23
	MsgMakeUp               = 24
	MsgInterruptRequest     = 25
	MsgTextMessage          = 27
	MsgPeripheralStatus     = 28
	MsgReset                = 30
	MsgIncidentNotification = 31
	MsgAlertCrew            = 40
	MsgAlertStatus          = 42
	MsgAlertEng             = 43
	MsgACK                  = 50
	MsgNAK                  = 51
	MsgSetParameter         = 60
	MsgParameterRequest     = 61
	MsgParameter            = 62
	MsgParamReqMultiple     = 63
	MsgTest                 = 64
	MsgPrinterStatus        = 65
	MsgMTAStatusChange      = 66
	MsgRouteStatus          = 67
	MsgSupplierMessage      = 100
	MsgBrigadeMessage       = 101
	MsgDataBaseQuery        = 102
	MsgFormattedText        = 103
	MsgProformaDefQuery     = 104
	MsgProformaDefinition   = 105
)

// Message is implemented by all GD92 message types.
type Message interface {
	// Type returns the message type number.
	Type() uint8
	// MarshalGD92 encodes the message body (not including the envelope).
	MarshalGD92() ([]byte, error)
	// UnmarshalGD92 decodes the message body.
	UnmarshalGD92(data []byte) error
}

// messageRegistry maps message type numbers to factory functions.
var messageRegistry = map[uint8]func() Message{}

// RegisterMessage registers a message type factory. Called from init() in messages.go.
func RegisterMessage(msgType uint8, factory func() Message) {
	messageRegistry[msgType] = factory
}

// ParseMessage creates and decodes a message from its type number and body data.
// Returns an error if the message type is unknown.
func ParseMessage(msgType uint8, data []byte) (Message, error) {
	factory, ok := messageRegistry[msgType]
	if !ok {
		return nil, fmt.Errorf("gd92: unknown message type %d", msgType)
	}
	msg := factory()
	if err := msg.UnmarshalGD92(data); err != nil {
		return nil, fmt.Errorf("gd92: unmarshal message type %d: %w", msgType, err)
	}
	return msg, nil
}

// ParseEnvelopeMessage parses the Message field of an envelope into a typed Message.
func ParseEnvelopeMessage(env *Envelope) (Message, error) {
	return ParseMessage(env.MessageType, env.Message)
}
