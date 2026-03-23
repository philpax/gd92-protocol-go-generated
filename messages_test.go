package gd92

import (
	"bytes"
	"testing"
)

// roundTripMessage is a helper that marshals then unmarshals and returns the result.
func roundTripMessage(t *testing.T, msg Message) Message {
	t.Helper()
	data, err := msg.MarshalGD92()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := ParseMessage(msg.Type(), data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func TestMobiliseCommandRoundTrip(t *testing.T) {
	msg := &MobiliseCommand{OpPeripherals: 0x00FF, ManAckReq: true}
	got := roundTripMessage(t, msg).(*MobiliseCommand)
	if got.OpPeripherals != msg.OpPeripherals {
		t.Fatalf("OpPeripherals: expected 0x%04x, got 0x%04x", msg.OpPeripherals, got.OpPeripherals)
	}
	if got.ManAckReq != msg.ManAckReq {
		t.Fatalf("ManAckReq: expected %v, got %v", msg.ManAckReq, got.ManAckReq)
	}
}

func TestMobiliseMessageRoundTrip(t *testing.T) {
	msg := &MobiliseMessage{
		Block:       1,
		OfBlocks:    2,
		ManAckReq:   true,
		TimeAndDate: "15MAR96143022",
		Callsigns:   []string{"E21", "E22"},
		Incidents: []MobiliseIncident{
			{
				IncidentNumber:  12345,
				MobilisationType: 1,
				Address: Address{
					AddressText: "Zone A",
					HouseNumber: "42",
					Street:      "HIGH STREET",
					SubDistrict: "",
					District:    "ANYTOWN",
					Town:        "SOMECITY",
					County:      "COUNTYSHIRE",
					Postcode:    "AB1 2CD",
				},
				MapRef:    "SU123456",
				TelNumber: "01onal234567",
				Text:      "PERSONS REPORTED",
			},
		},
	}
	got := roundTripMessage(t, msg).(*MobiliseMessage)
	if got.Block != msg.Block {
		t.Fatalf("Block: expected %d, got %d", msg.Block, got.Block)
	}
	if got.OfBlocks != msg.OfBlocks {
		t.Fatalf("OfBlocks: expected %d, got %d", msg.OfBlocks, got.OfBlocks)
	}
	if got.ManAckReq != msg.ManAckReq {
		t.Fatal("ManAckReq mismatch")
	}
	if got.TimeAndDate != msg.TimeAndDate {
		t.Fatalf("TimeAndDate: expected %q, got %q", msg.TimeAndDate, got.TimeAndDate)
	}
	if len(got.Callsigns) != 2 || got.Callsigns[0] != "E21" || got.Callsigns[1] != "E22" {
		t.Fatalf("Callsigns: expected [E21 E22], got %v", got.Callsigns)
	}
	if len(got.Incidents) != 1 {
		t.Fatalf("expected 1 incident, got %d", len(got.Incidents))
	}
	inc := got.Incidents[0]
	if inc.IncidentNumber != 12345 {
		t.Fatalf("IncidentNumber: expected 12345, got %d", inc.IncidentNumber)
	}
	if inc.Address.Street != "HIGH STREET" {
		t.Fatalf("Street: expected HIGH STREET, got %q", inc.Address.Street)
	}
	if inc.Text != "PERSONS REPORTED" {
		t.Fatalf("Text: expected PERSONS REPORTED, got %q", inc.Text)
	}
}

func TestPageOfficerRoundTrip(t *testing.T) {
	msg := &PageOfficer{
		PagerPriority: 'E',
		PagerNumber:   PagerNumber{TelNumber: "07700900123", PagerType: 'A'},
		PagerText:     "FIRE ALARM ZONE 3",
	}
	got := roundTripMessage(t, msg).(*PageOfficer)
	if got.PagerPriority != 'E' {
		t.Fatalf("PagerPriority: expected 'E', got %c", got.PagerPriority)
	}
	if got.PagerNumber.TelNumber != "07700900123" {
		t.Fatalf("TelNumber: expected 07700900123, got %q", got.PagerNumber.TelNumber)
	}
	if got.PagerNumber.PagerType != 'A' {
		t.Fatalf("PagerType: expected 'A', got %c", got.PagerNumber.PagerType)
	}
	if got.PagerText != "FIRE ALARM ZONE 3" {
		t.Fatalf("PagerText: expected %q, got %q", msg.PagerText, got.PagerText)
	}
}

func TestResourceStatusRoundTrip(t *testing.T) {
	msg := &ResourceStatus{
		Resources: []ResourceEntry{
			{
				Callsign:   "E21",
				AVLType:    0,
				AVLData:    "",
				StatusCode: 4, // Available at Base
				Remarks:    "ALL OK",
			},
			{
				Callsign:   "E22",
				AVLType:    1,
				AVLData:    "GPS DATA HERE",
				StatusCode: 1, // Mobile to Incident
				Remarks:    "EN ROUTE",
			},
		},
	}
	got := roundTripMessage(t, msg).(*ResourceStatus)
	if len(got.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(got.Resources))
	}
	if got.Resources[0].Callsign != "E21" {
		t.Fatalf("Resource[0].Callsign: expected E21, got %q", got.Resources[0].Callsign)
	}
	if got.Resources[0].StatusCode != 4 {
		t.Fatalf("Resource[0].StatusCode: expected 4, got %d", got.Resources[0].StatusCode)
	}
	if got.Resources[1].AVLData != "GPS DATA HERE" {
		t.Fatalf("Resource[1].AVLData: expected %q, got %q", "GPS DATA HERE", got.Resources[1].AVLData)
	}
}

func TestTextMessageRoundTrip(t *testing.T) {
	msg := &TextMessage{
		Block:    1,
		OfBlocks: 1,
		Text:     "THIS IS A TEST MESSAGE WITH SPACES          THAT SHOULD COMPRESS",
	}
	got := roundTripMessage(t, msg).(*TextMessage)
	if got.Text != msg.Text {
		t.Fatalf("Text: expected %q, got %q", msg.Text, got.Text)
	}
}

func TestIncidentNotificationRoundTrip(t *testing.T) {
	msg := &IncidentNotification{
		AlarmType:   "FIRE",
		CallAgency:  "ADT",
		TelNumber:   "020712345",
		AlarmRef:    100,
		AlarmSerial: "SER001",
		Address: Address{
			AddressText: "TESCO EXPRESS",
			HouseNumber: "1",
			Street:      "MAIN ROAD",
			Town:        "LONDON",
			Postcode:    "SE1 1AA",
		},
		Text: "SMOKE DETECTOR ZONE 2",
	}
	got := roundTripMessage(t, msg).(*IncidentNotification)
	if got.AlarmType != "FIRE" {
		t.Fatalf("AlarmType: expected FIRE, got %q", got.AlarmType)
	}
	if got.AlarmRef != 100 {
		t.Fatalf("AlarmRef: expected 100, got %d", got.AlarmRef)
	}
	if got.Address.Street != "MAIN ROAD" {
		t.Fatalf("Street: expected MAIN ROAD, got %q", got.Address.Street)
	}
	if got.Text != "SMOKE DETECTOR ZONE 2" {
		t.Fatalf("Text: expected %q, got %q", msg.Text, got.Text)
	}
}

func TestACKRoundTrip(t *testing.T) {
	msg := &ACK{}
	got := roundTripMessage(t, msg).(*ACK)
	_ = got // ACK has no fields
}

func TestNAKRoundTrip(t *testing.T) {
	msg := &NAK{
		Count:         1,
		Destinations:  []CommsAddress{{Brigade: 5, Node: 42, Port: 3}},
		ReasonCodeSet: ReasonSetGeneral,
		ReasonCode:    GeneralCheckError,
	}
	got := roundTripMessage(t, msg).(*NAK)
	if got.Count != 1 {
		t.Fatalf("Count: expected 1, got %d", got.Count)
	}
	if !got.Destinations[0].Equal(msg.Destinations[0]) {
		t.Fatalf("Destinations[0]: expected %v, got %v", msg.Destinations[0], got.Destinations[0])
	}
	if got.ReasonCodeSet != ReasonSetGeneral {
		t.Fatalf("ReasonCodeSet: expected %d, got %d", ReasonSetGeneral, got.ReasonCodeSet)
	}
	if got.ReasonCode != GeneralCheckError {
		t.Fatalf("ReasonCode: expected %d, got %d", GeneralCheckError, got.ReasonCode)
	}
}

func TestStopRoundTrip(t *testing.T) {
	msg := &Stop{
		Callsign:       "E21",
		IncidentNumber: 99999,
		StopCode:       "STOP1",
	}
	got := roundTripMessage(t, msg).(*Stop)
	if got.StopCode != "STOP1" {
		t.Fatalf("StopCode: expected STOP1, got %q", got.StopCode)
	}
}

func TestMakeUpRoundTrip(t *testing.T) {
	msg := &MakeUp{
		Callsign:       "E21",
		IncidentNumber: 50000,
		NumberTypes:    2,
		Appliances: []ApplianceRequest{
			{ApplType: "HP", ApplQuantity: 1},
			{ApplType: "FEV", ApplQuantity: 2},
		},
	}
	got := roundTripMessage(t, msg).(*MakeUp)
	if len(got.Appliances) != 2 {
		t.Fatalf("expected 2 appliances, got %d", len(got.Appliances))
	}
	if got.Appliances[0].ApplType != "HP" {
		t.Fatalf("Appliance[0]: expected HP, got %q", got.Appliances[0].ApplType)
	}
	if got.Appliances[1].ApplType != "FEV" {
		t.Fatalf("Appliance[1]: expected FEV, got %q", got.Appliances[1].ApplType)
	}
}

func TestSetParameterRoundTrip(t *testing.T) {
	msg := &SetParameter{
		ParameterTable: 1,
		ParameterNo:    5,
		Entries: []ParameterEntry{
			{Index: 0, Value: []byte{0x42}},
		},
	}
	got := roundTripMessage(t, msg).(*SetParameter)
	if got.ParameterTable != 1 {
		t.Fatalf("ParameterTable: expected 1, got %d", got.ParameterTable)
	}
	if got.ParameterNo != 5 {
		t.Fatalf("ParameterNo: expected 5, got %d", got.ParameterNo)
	}
	if len(got.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got.Entries))
	}
	if got.Entries[0].Index != 0 {
		t.Fatalf("Index: expected 0, got %d", got.Entries[0].Index)
	}
}

func TestMTAStatusChangeRoundTrip(t *testing.T) {
	msg := &MTAStatusChange{MTAStatus: 1}
	got := roundTripMessage(t, msg).(*MTAStatusChange)
	if got.MTAStatus != 1 {
		t.Fatalf("MTAStatus: expected 1, got %d", got.MTAStatus)
	}
}

func TestRouteStatusRoundTrip(t *testing.T) {
	msg := &RouteStatus{
		Enabled: true,
		DestinationNodes: []AddressRange{
			{
				Start: CommsAddress{Brigade: 1, Node: 0, Port: 0},
				End:   CommsAddress{Brigade: 1, Node: 10, Port: 63},
			},
		},
	}
	got := roundTripMessage(t, msg).(*RouteStatus)
	if !got.Enabled {
		t.Fatal("expected Enabled=true")
	}
	if len(got.DestinationNodes) != 1 {
		t.Fatalf("expected 1 range, got %d", len(got.DestinationNodes))
	}
	if !got.DestinationNodes[0].Start.Equal(msg.DestinationNodes[0].Start) {
		t.Fatalf("Start mismatch")
	}
}

func TestAlertCrewRoundTrip(t *testing.T) {
	msg := &AlertCrew{
		AlertGroup:    "FA",
		ManAckReq:     true,
		OpPeripherals: 0x0003,
	}
	got := roundTripMessage(t, msg).(*AlertCrew)
	if got.AlertGroup != "FA" {
		t.Fatalf("AlertGroup: expected FA, got %q", got.AlertGroup)
	}
	if !got.ManAckReq {
		t.Fatal("expected ManAckReq=true")
	}
}

func TestAlertStatusRoundTrip(t *testing.T) {
	// 1-byte status
	msg := &AlertStatus{AlerterStatus: "A"}
	got := roundTripMessage(t, msg).(*AlertStatus)
	if got.AlerterStatus != "A" {
		t.Fatalf("AlerterStatus: expected A, got %q", got.AlerterStatus)
	}

	// 2-byte status
	msg2 := &AlertStatus{AlerterStatus: "ba"}
	got2 := roundTripMessage(t, msg2).(*AlertStatus)
	if got2.AlerterStatus != "ba" {
		t.Fatalf("AlerterStatus: expected ba, got %q", got2.AlerterStatus)
	}
}

func TestPeripheralStatusRoundTrip(t *testing.T) {
	msg := &PeripheralStatus{IpPeripherals: 0x1234, OpPeripherals: 0x5678}
	got := roundTripMessage(t, msg).(*PeripheralStatus)
	if got.IpPeripherals != 0x1234 {
		t.Fatalf("IpPeripherals: expected 0x1234, got 0x%04x", got.IpPeripherals)
	}
	if got.OpPeripherals != 0x5678 {
		t.Fatalf("OpPeripherals: expected 0x5678, got 0x%04x", got.OpPeripherals)
	}
}

func TestResetRoundTrip(t *testing.T) {
	msg := &Reset{ResetReason: 2}
	got := roundTripMessage(t, msg).(*Reset)
	if got.ResetReason != 2 {
		t.Fatalf("ResetReason: expected 2, got %d", got.ResetReason)
	}
}

func TestSupplierMessageRoundTrip(t *testing.T) {
	msg := &SupplierMessage{Data: []byte{0xDE, 0xAD, 0xBE, 0xEF}}
	got := roundTripMessage(t, msg).(*SupplierMessage)
	if !bytes.Equal(got.Data, msg.Data) {
		t.Fatalf("Data mismatch")
	}
}

func TestBrigadeMessageRoundTrip(t *testing.T) {
	msg := &BrigadeMessage{Text: "ADMIN MESSAGE WITH      LOTS OF SPACES"}
	got := roundTripMessage(t, msg).(*BrigadeMessage)
	if got.Text != msg.Text {
		t.Fatalf("Text: expected %q, got %q", msg.Text, got.Text)
	}
}

func TestLogUpdateRoundTrip(t *testing.T) {
	msg := &LogUpdate{
		Callsign:       "E21",
		IncidentNumber: 42,
		Update:         "FIRE UNDER CONTROL",
	}
	got := roundTripMessage(t, msg).(*LogUpdate)
	if got.Callsign != "E21" {
		t.Fatalf("Callsign: expected E21, got %q", got.Callsign)
	}
	if got.IncidentNumber != 42 {
		t.Fatalf("IncidentNumber: expected 42, got %d", got.IncidentNumber)
	}
	if got.Update != "FIRE UNDER CONTROL" {
		t.Fatalf("Update: expected %q, got %q", msg.Update, got.Update)
	}
}

func TestInterruptRequestRoundTrip(t *testing.T) {
	msg := &InterruptRequest{
		Callsign:    "E21",
		RequestCode: 'E',
		Text:        "EMERGENCY",
	}
	got := roundTripMessage(t, msg).(*InterruptRequest)
	if got.RequestCode != 'E' {
		t.Fatalf("RequestCode: expected 'E', got %c", got.RequestCode)
	}
}

func TestDutyStaffingUpdateRoundTrip(t *testing.T) {
	msg := &DutyStaffingUpdate{
		Entries: []DutyStaffingEntry{
			{
				Callsign:   "E21",
				OIC:        "SM JONES",
				Riders:     5,
				StatusCode: 4,
				Remarks:    "FULL CREW",
			},
		},
	}
	got := roundTripMessage(t, msg).(*DutyStaffingUpdate)
	if len(got.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got.Entries))
	}
	if got.Entries[0].OIC != "SM JONES" {
		t.Fatalf("OIC: expected SM JONES, got %q", got.Entries[0].OIC)
	}
	if got.Entries[0].Riders != 5 {
		t.Fatalf("Riders: expected 5, got %d", got.Entries[0].Riders)
	}
}

func TestParamReqMultipleRoundTrip(t *testing.T) {
	msg := &ParamReqMultiple{
		ParameterTable: 0,
		ParameterNo:    13,
		FirstEntry:     0,
		LastEntry:      10,
	}
	got := roundTripMessage(t, msg).(*ParamReqMultiple)
	if got.ParameterNo != 13 {
		t.Fatalf("ParameterNo: expected 13, got %d", got.ParameterNo)
	}
	if got.LastEntry != 10 {
		t.Fatalf("LastEntry: expected 10, got %d", got.LastEntry)
	}
}

func TestUnknownMessageType(t *testing.T) {
	_, err := ParseMessage(255, []byte{0x00})
	if err == nil {
		t.Fatal("expected error for unknown message type")
	}
}

// TestFullEnvelopeWithMessage tests a complete envelope containing a typed message.
func TestFullEnvelopeWithMessage(t *testing.T) {
	msg := &MobiliseCommand{OpPeripherals: 0x0003, ManAckReq: true}
	msgData, err := msg.MarshalGD92()
	if err != nil {
		t.Fatal(err)
	}

	env := &Envelope{
		Source:       CommsAddress{Brigade: 1, Node: 0, Port: 1},
		Destinations: []CommsAddress{{Brigade: 1, Node: 10, Port: 4}},
		Priority:     1,
		ProtVers:     1,
		AckReq:       true,
		Seq:          1,
		MessageType:  MsgMobiliseCommand,
		Message:      msgData,
	}

	wireData, err := env.MarshalEnvelope()
	if err != nil {
		t.Fatal(err)
	}

	// Wrap in frame
	frame := WrapFrame(wireData)

	// Unwrap frame
	unwrapped := UnwrapFrame(frame)
	if unwrapped == nil {
		t.Fatal("failed to unwrap frame")
	}

	// Unmarshal envelope
	gotEnv, err := UnmarshalEnvelope(unwrapped)
	if err != nil {
		t.Fatal(err)
	}

	// Parse message
	gotMsg, err := ParseEnvelopeMessage(gotEnv)
	if err != nil {
		t.Fatal(err)
	}

	mc, ok := gotMsg.(*MobiliseCommand)
	if !ok {
		t.Fatalf("expected *MobiliseCommand, got %T", gotMsg)
	}
	if mc.OpPeripherals != 0x0003 {
		t.Fatalf("OpPeripherals: expected 0x0003, got 0x%04x", mc.OpPeripherals)
	}
	if !mc.ManAckReq {
		t.Fatal("expected ManAckReq=true")
	}
}
