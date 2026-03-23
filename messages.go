package gd92

import "strings"

func init() {
	RegisterMessage(MsgMobiliseCommand, func() Message { return &MobiliseCommand{} })
	RegisterMessage(MsgMobiliseMessage, func() Message { return &MobiliseMessage{} })
	RegisterMessage(MsgPageOfficer, func() Message { return &PageOfficer{} })
	RegisterMessage(MsgAreaPageMessage, func() Message { return &AreaPageMessage{} })
	RegisterMessage(MsgResourceStatusReq, func() Message { return &ResourceStatusRequest{} })
	RegisterMessage(MsgMobiliseMessageAlt, func() Message { return &MobiliseMessageAlt{} })
	RegisterMessage(MsgActivatePeripheral, func() Message { return &ActivatePeripheral{} })
	RegisterMessage(MsgDeactivatePeripheral, func() Message { return &DeactivatePeripheral{} })
	RegisterMessage(MsgPeripheralStatusReq, func() Message { return &PeripheralStatusRequest{} })
	RegisterMessage(MsgResetRequest, func() Message { return &ResetRequest{} })
	RegisterMessage(MsgResourceStatus, func() Message { return &ResourceStatus{} })
	RegisterMessage(MsgDutyStaffingUpdate, func() Message { return &DutyStaffingUpdate{} })
	RegisterMessage(MsgLogUpdate, func() Message { return &LogUpdate{} })
	RegisterMessage(MsgStop, func() Message { return &Stop{} })
	RegisterMessage(MsgMakeUp, func() Message { return &MakeUp{} })
	RegisterMessage(MsgInterruptRequest, func() Message { return &InterruptRequest{} })
	RegisterMessage(MsgTextMessage, func() Message { return &TextMessage{} })
	RegisterMessage(MsgPeripheralStatus, func() Message { return &PeripheralStatus{} })
	RegisterMessage(MsgReset, func() Message { return &Reset{} })
	RegisterMessage(MsgIncidentNotification, func() Message { return &IncidentNotification{} })
	RegisterMessage(MsgAlertCrew, func() Message { return &AlertCrew{} })
	RegisterMessage(MsgAlertStatus, func() Message { return &AlertStatus{} })
	RegisterMessage(MsgAlertEng, func() Message { return &AlertEng{} })
	RegisterMessage(MsgACK, func() Message { return &ACK{} })
	RegisterMessage(MsgNAK, func() Message { return &NAK{} })
	RegisterMessage(MsgSetParameter, func() Message { return &SetParameter{} })
	RegisterMessage(MsgParameterRequest, func() Message { return &ParameterRequest{} })
	RegisterMessage(MsgParameter, func() Message { return &Parameter{} })
	RegisterMessage(MsgParamReqMultiple, func() Message { return &ParamReqMultiple{} })
	RegisterMessage(MsgTest, func() Message { return &Test{} })
	RegisterMessage(MsgPrinterStatus, func() Message { return &PrinterStatusMsg{} })
	RegisterMessage(MsgMTAStatusChange, func() Message { return &MTAStatusChange{} })
	RegisterMessage(MsgRouteStatus, func() Message { return &RouteStatus{} })
	RegisterMessage(MsgSupplierMessage, func() Message { return &SupplierMessage{} })
	RegisterMessage(MsgBrigadeMessage, func() Message { return &BrigadeMessage{} })
	RegisterMessage(MsgDataBaseQuery, func() Message { return &DataBaseQuery{} })
	RegisterMessage(MsgFormattedText, func() Message { return &FormattedText{} })
	RegisterMessage(MsgProformaDefQuery, func() Message { return &ProformaDefQuery{} })
	RegisterMessage(MsgProformaDefinition, func() Message { return &ProformaDefinition{} })
}

// Address represents a composite address as defined in spec A.4.
type Address struct {
	AddressText string // compressed_string, 0-120
	HouseNumber string // string, 0-10
	Street      string // compressed_string, 0-40
	SubDistrict string // compressed_string, 0-30
	District    string // compressed_string, 0-30
	Town        string // compressed_string, 0-30
	County      string // compressed_string, 0-20
	Postcode    string // string, 0-10
}

func (a *Address) marshal(enc *Encoder) {
	enc.WriteCompressedString(a.AddressText)
	enc.WriteString(a.HouseNumber)
	enc.WriteCompressedString(a.Street)
	enc.WriteCompressedString(a.SubDistrict)
	enc.WriteCompressedString(a.District)
	enc.WriteCompressedString(a.Town)
	enc.WriteCompressedString(a.County)
	enc.WriteString(a.Postcode)
}

func (a *Address) unmarshal(d *Decoder) error {
	var err error
	if a.AddressText, err = d.ReadCompressedString(); err != nil {
		return err
	}
	if a.HouseNumber, err = d.ReadString(); err != nil {
		return err
	}
	if a.Street, err = d.ReadCompressedString(); err != nil {
		return err
	}
	if a.SubDistrict, err = d.ReadCompressedString(); err != nil {
		return err
	}
	if a.District, err = d.ReadCompressedString(); err != nil {
		return err
	}
	if a.Town, err = d.ReadCompressedString(); err != nil {
		return err
	}
	if a.County, err = d.ReadCompressedString(); err != nil {
		return err
	}
	if a.Postcode, err = d.ReadString(); err != nil {
		return err
	}
	return nil
}

// PagerNumber is <tel_number><pager_type>.
type PagerNumber struct {
	TelNumber string // string, 0-16
	PagerType byte   // ASCII: 'A'=alpha, 'N'=numeric, 'T'=tones
}

// MobiliseIncident represents one repeating incident group in Mobilise_message.
type MobiliseIncident struct {
	IncidentNumber  uint32
	MobilisationType uint8
	Address          Address
	MapRef           string // string, 0-16
	TelNumber        string // string, 0-16
	Text             string // long_comp_string
}

// ResourceEntry represents one repeating resource group in Resource_status.
type ResourceEntry struct {
	Callsign   string // string, 0-6
	AVLType    uint8
	AVLData    string // string, 0-40
	StatusCode uint8
	Remarks    string // compressed_string, 0-200
}

// DutyStaffingEntry represents one repeating group in Duty_staffing_update.
type DutyStaffingEntry struct {
	Callsign   string // string, 0-6
	OIC        string // string, 0-20
	Riders     uint8  // 1-15
	StatusCode uint8
	Remarks    string // compressed_string, 0-200
}

// ApplianceRequest represents one repeating group in Make-up.
type ApplianceRequest struct {
	ApplType     string // 3 bytes fixed ASCII
	ApplQuantity uint8
}

// ParameterEntry represents one repeating group in Set_parameter.
type ParameterEntry struct {
	Index uint16
	Value []byte // raw parameter value bytes
}

// AddressRange is <comms_address><comms_address>.
type AddressRange struct {
	Start CommsAddress
	End   CommsAddress
}

// --- Message 01: Mobilise_command ---

type MobiliseCommand struct {
	OpPeripherals uint16
	ManAckReq     bool
}

func (m *MobiliseCommand) Type() uint8 { return MsgMobiliseCommand }

func (m *MobiliseCommand) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord16(m.OpPeripherals)
	enc.WriteBool(m.ManAckReq)
	return enc.Bytes(), nil
}

func (m *MobiliseCommand) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.OpPeripherals, err = d.ReadWord16(); err != nil {
		return err
	}
	if m.ManAckReq, err = d.ReadBool(); err != nil {
		return err
	}
	return nil
}

// --- Message 02: Mobilise_message ---

type MobiliseMessage struct {
	Block     uint8
	OfBlocks  uint8
	ManAckReq bool
	TimeAndDate string // 13 bytes fixed
	Callsigns   []string // callsign_list: count + callsigns
	Incidents   []MobiliseIncident
}

func (m *MobiliseMessage) Type() uint8 { return MsgMobiliseMessage }

func (m *MobiliseMessage) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.Block)
	enc.WriteWord8(m.OfBlocks)
	enc.WriteBool(m.ManAckReq)
	enc.WriteTimeAndDate(m.TimeAndDate)
	// callsign_list: count followed by callsigns
	enc.WriteWord8(uint8(len(m.Callsigns)))
	for _, cs := range m.Callsigns {
		enc.WriteString(cs)
	}
	// repeating incidents
	for _, inc := range m.Incidents {
		enc.WriteWord32(inc.IncidentNumber)
		enc.WriteWord8(inc.MobilisationType)
		inc.Address.marshal(enc)
		enc.WriteString(inc.MapRef)
		enc.WriteString(inc.TelNumber)
		enc.WriteLongCompString(inc.Text)
	}
	return enc.Bytes(), nil
}

func (m *MobiliseMessage) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Block, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.OfBlocks, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.ManAckReq, err = d.ReadBool(); err != nil {
		return err
	}
	if m.TimeAndDate, err = d.ReadTimeAndDate(); err != nil {
		return err
	}
	// callsign_list
	csCount, err := d.ReadWord8()
	if err != nil {
		return err
	}
	m.Callsigns = make([]string, csCount)
	for i := range m.Callsigns {
		if m.Callsigns[i], err = d.ReadString(); err != nil {
			return err
		}
	}
	// repeating incidents
	for d.Remaining() > 0 {
		var inc MobiliseIncident
		if inc.IncidentNumber, err = d.ReadWord32(); err != nil {
			return err
		}
		if inc.MobilisationType, err = d.ReadWord8(); err != nil {
			return err
		}
		if err = inc.Address.unmarshal(d); err != nil {
			return err
		}
		if inc.MapRef, err = d.ReadString(); err != nil {
			return err
		}
		if inc.TelNumber, err = d.ReadString(); err != nil {
			return err
		}
		if inc.Text, err = d.ReadLongCompString(); err != nil {
			return err
		}
		m.Incidents = append(m.Incidents, inc)
	}
	return nil
}

// --- Message 03: Page_officer ---

type PageOfficer struct {
	PagerPriority byte // ASCII: E/P/R/A
	PagerNumber   PagerNumber
	PagerText     string // compressed_string, 0-200
}

func (m *PageOfficer) Type() uint8 { return MsgPageOfficer }

func (m *PageOfficer) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.PagerPriority)
	enc.WriteString(m.PagerNumber.TelNumber)
	enc.WriteWord8(m.PagerNumber.PagerType)
	enc.WriteCompressedString(m.PagerText)
	return enc.Bytes(), nil
}

func (m *PageOfficer) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.PagerPriority, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.PagerNumber.TelNumber, err = d.ReadString(); err != nil {
		return err
	}
	if m.PagerNumber.PagerType, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.PagerText, err = d.ReadCompressedString(); err != nil {
		return err
	}
	return nil
}

// --- Message 04: Area_page_message ---

type AreaPageMessage struct {
	PagerPriority byte
	PagerNumber   PagerNumber
	PagerText     string
}

func (m *AreaPageMessage) Type() uint8 { return MsgAreaPageMessage }

func (m *AreaPageMessage) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.PagerPriority)
	enc.WriteString(m.PagerNumber.TelNumber)
	enc.WriteWord8(m.PagerNumber.PagerType)
	enc.WriteCompressedString(m.PagerText)
	return enc.Bytes(), nil
}

func (m *AreaPageMessage) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.PagerPriority, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.PagerNumber.TelNumber, err = d.ReadString(); err != nil {
		return err
	}
	if m.PagerNumber.PagerType, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.PagerText, err = d.ReadCompressedString(); err != nil {
		return err
	}
	return nil
}

// --- Message 05: Resource_status_request ---

type ResourceStatusRequest struct {
	Callsigns []string
}

func (m *ResourceStatusRequest) Type() uint8 { return MsgResourceStatusReq }

func (m *ResourceStatusRequest) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	for _, cs := range m.Callsigns {
		enc.WriteString(cs)
	}
	return enc.Bytes(), nil
}

func (m *ResourceStatusRequest) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	for d.Remaining() > 0 {
		cs, err := d.ReadString()
		if err != nil {
			return err
		}
		m.Callsigns = append(m.Callsigns, cs)
	}
	return nil
}

// --- Message 06: Mobilise_message (Hampshire/Kent) ---
// Format undefined per spec B.34 - stored as raw bytes.

type MobiliseMessageAlt struct {
	RawData []byte
}

func (m *MobiliseMessageAlt) Type() uint8 { return MsgMobiliseMessageAlt }

func (m *MobiliseMessageAlt) MarshalGD92() ([]byte, error) {
	out := make([]byte, len(m.RawData))
	copy(out, m.RawData)
	return out, nil
}

func (m *MobiliseMessageAlt) UnmarshalGD92(data []byte) error {
	m.RawData = make([]byte, len(data))
	copy(m.RawData, data)
	return nil
}

// --- Message 07: Activate_peripheral ---

type ActivatePeripheral struct {
	OpPeripherals uint16
}

func (m *ActivatePeripheral) Type() uint8 { return MsgActivatePeripheral }

func (m *ActivatePeripheral) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord16(m.OpPeripherals)
	return enc.Bytes(), nil
}

func (m *ActivatePeripheral) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.OpPeripherals, err = d.ReadWord16()
	return err
}

// --- Message 08: Deactivate_peripheral ---

type DeactivatePeripheral struct {
	OpPeripherals uint16
}

func (m *DeactivatePeripheral) Type() uint8 { return MsgDeactivatePeripheral }

func (m *DeactivatePeripheral) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord16(m.OpPeripherals)
	return enc.Bytes(), nil
}

func (m *DeactivatePeripheral) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.OpPeripherals, err = d.ReadWord16()
	return err
}

// --- Message 09: Peripheral_status_request ---

type PeripheralStatusRequest struct{}

func (m *PeripheralStatusRequest) Type() uint8 { return MsgPeripheralStatusReq }

func (m *PeripheralStatusRequest) MarshalGD92() ([]byte, error) {
	return nil, nil
}

func (m *PeripheralStatusRequest) UnmarshalGD92(data []byte) error {
	return nil
}

// --- Message 10: Reset_request ---

type ResetRequest struct {
	ResetType uint8
}

func (m *ResetRequest) Type() uint8 { return MsgResetRequest }

func (m *ResetRequest) MarshalGD92() ([]byte, error) {
	return []byte{m.ResetType}, nil
}

func (m *ResetRequest) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.ResetType, err = d.ReadWord8()
	return err
}

// --- Message 20: Resource_status ---

type ResourceStatus struct {
	Resources []ResourceEntry
}

func (m *ResourceStatus) Type() uint8 { return MsgResourceStatus }

func (m *ResourceStatus) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	for _, r := range m.Resources {
		enc.WriteString(r.Callsign)
		enc.WriteWord8(r.AVLType)
		enc.WriteString(r.AVLData)
		enc.WriteWord8(r.StatusCode)
		enc.WriteCompressedString(r.Remarks)
	}
	return enc.Bytes(), nil
}

func (m *ResourceStatus) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	for d.Remaining() > 0 {
		var r ResourceEntry
		var err error
		if r.Callsign, err = d.ReadString(); err != nil {
			return err
		}
		if r.AVLType, err = d.ReadWord8(); err != nil {
			return err
		}
		if r.AVLData, err = d.ReadString(); err != nil {
			return err
		}
		if r.StatusCode, err = d.ReadWord8(); err != nil {
			return err
		}
		if r.Remarks, err = d.ReadCompressedString(); err != nil {
			return err
		}
		m.Resources = append(m.Resources, r)
	}
	return nil
}

// --- Message 21: Duty_staffing_update ---

type DutyStaffingUpdate struct {
	Entries []DutyStaffingEntry
}

func (m *DutyStaffingUpdate) Type() uint8 { return MsgDutyStaffingUpdate }

func (m *DutyStaffingUpdate) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	for _, e := range m.Entries {
		enc.WriteString(e.Callsign)
		enc.WriteString(e.OIC)
		enc.WriteWord8(e.Riders)
		enc.WriteWord8(e.StatusCode)
		enc.WriteCompressedString(e.Remarks)
	}
	return enc.Bytes(), nil
}

func (m *DutyStaffingUpdate) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	for d.Remaining() > 0 {
		var e DutyStaffingEntry
		var err error
		if e.Callsign, err = d.ReadString(); err != nil {
			return err
		}
		if e.OIC, err = d.ReadString(); err != nil {
			return err
		}
		if e.Riders, err = d.ReadWord8(); err != nil {
			return err
		}
		if e.StatusCode, err = d.ReadWord8(); err != nil {
			return err
		}
		if e.Remarks, err = d.ReadCompressedString(); err != nil {
			return err
		}
		m.Entries = append(m.Entries, e)
	}
	return nil
}

// --- Message 22: Log_update ---

type LogUpdate struct {
	Callsign       string
	IncidentNumber uint32
	Update         string // compressed_string
}

func (m *LogUpdate) Type() uint8 { return MsgLogUpdate }

func (m *LogUpdate) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteString(m.Callsign)
	enc.WriteWord32(m.IncidentNumber)
	enc.WriteCompressedString(m.Update)
	return enc.Bytes(), nil
}

func (m *LogUpdate) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Callsign, err = d.ReadString(); err != nil {
		return err
	}
	if m.IncidentNumber, err = d.ReadWord32(); err != nil {
		return err
	}
	if m.Update, err = d.ReadCompressedString(); err != nil {
		return err
	}
	return nil
}

// --- Message 23: Stop ---

type Stop struct {
	Callsign       string
	IncidentNumber uint32
	StopCode       string // 5 bytes fixed ASCII
}

func (m *Stop) Type() uint8 { return MsgStop }

func (m *Stop) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteString(m.Callsign)
	enc.WriteWord32(m.IncidentNumber)
	enc.WriteFixedASCII(m.StopCode, 5)
	return enc.Bytes(), nil
}

func (m *Stop) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Callsign, err = d.ReadString(); err != nil {
		return err
	}
	if m.IncidentNumber, err = d.ReadWord32(); err != nil {
		return err
	}
	if m.StopCode, err = d.ReadFixedASCII(5); err != nil {
		return err
	}
	m.StopCode = strings.TrimRight(m.StopCode, " ")
	return nil
}

// --- Message 24: Make-up ---

type MakeUp struct {
	Callsign       string
	IncidentNumber uint32
	NumberTypes    uint8
	Appliances     []ApplianceRequest
}

func (m *MakeUp) Type() uint8 { return MsgMakeUp }

func (m *MakeUp) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteString(m.Callsign)
	enc.WriteWord32(m.IncidentNumber)
	enc.WriteWord8(m.NumberTypes)
	for _, a := range m.Appliances {
		enc.WriteFixedASCII(a.ApplType, 3)
		enc.WriteWord8(a.ApplQuantity)
	}
	return enc.Bytes(), nil
}

func (m *MakeUp) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Callsign, err = d.ReadString(); err != nil {
		return err
	}
	if m.IncidentNumber, err = d.ReadWord32(); err != nil {
		return err
	}
	if m.NumberTypes, err = d.ReadWord8(); err != nil {
		return err
	}
	for d.Remaining() > 0 {
		var a ApplianceRequest
		if a.ApplType, err = d.ReadFixedASCII(3); err != nil {
			return err
		}
		a.ApplType = strings.TrimRight(a.ApplType, " ")
		if a.ApplQuantity, err = d.ReadWord8(); err != nil {
			return err
		}
		m.Appliances = append(m.Appliances, a)
	}
	return nil
}

// --- Message 25: Interrupt_request ---

type InterruptRequest struct {
	Callsign    string
	RequestCode byte // ASCII: S/E/C
	Text        string // long_comp_string
}

func (m *InterruptRequest) Type() uint8 { return MsgInterruptRequest }

func (m *InterruptRequest) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteString(m.Callsign)
	enc.WriteWord8(m.RequestCode)
	enc.WriteLongCompString(m.Text)
	return enc.Bytes(), nil
}

func (m *InterruptRequest) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Callsign, err = d.ReadString(); err != nil {
		return err
	}
	if m.RequestCode, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.Text, err = d.ReadLongCompString(); err != nil {
		return err
	}
	return nil
}

// --- Message 27: Text_message ---

type TextMessage struct {
	Block    uint8
	OfBlocks uint8
	Text     string // long_comp_string
}

func (m *TextMessage) Type() uint8 { return MsgTextMessage }

func (m *TextMessage) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.Block)
	enc.WriteWord8(m.OfBlocks)
	enc.WriteLongCompString(m.Text)
	return enc.Bytes(), nil
}

func (m *TextMessage) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Block, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.OfBlocks, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.Text, err = d.ReadLongCompString(); err != nil {
		return err
	}
	return nil
}

// --- Message 28: Peripheral_status ---

type PeripheralStatus struct {
	IpPeripherals uint16
	OpPeripherals uint16
}

func (m *PeripheralStatus) Type() uint8 { return MsgPeripheralStatus }

func (m *PeripheralStatus) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord16(m.IpPeripherals)
	enc.WriteWord16(m.OpPeripherals)
	return enc.Bytes(), nil
}

func (m *PeripheralStatus) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.IpPeripherals, err = d.ReadWord16(); err != nil {
		return err
	}
	if m.OpPeripherals, err = d.ReadWord16(); err != nil {
		return err
	}
	return nil
}

// --- Message 30: Reset ---

type Reset struct {
	ResetReason uint8
}

func (m *Reset) Type() uint8 { return MsgReset }

func (m *Reset) MarshalGD92() ([]byte, error) {
	return []byte{m.ResetReason}, nil
}

func (m *Reset) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.ResetReason, err = d.ReadWord8()
	return err
}

// --- Message 31: Incident_notification ---

type IncidentNotification struct {
	AlarmType   string // string, 0-10
	CallAgency  string // string, 0-10
	TelNumber   string // string, 0-16
	AlarmRef    uint16
	AlarmSerial string // string, 0-12
	Address     Address
	Text        string // long_comp_string
}

func (m *IncidentNotification) Type() uint8 { return MsgIncidentNotification }

func (m *IncidentNotification) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteString(m.AlarmType)
	enc.WriteString(m.CallAgency)
	enc.WriteString(m.TelNumber)
	enc.WriteWord16(m.AlarmRef)
	enc.WriteString(m.AlarmSerial)
	m.Address.marshal(enc)
	enc.WriteLongCompString(m.Text)
	return enc.Bytes(), nil
}

func (m *IncidentNotification) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.AlarmType, err = d.ReadString(); err != nil {
		return err
	}
	if m.CallAgency, err = d.ReadString(); err != nil {
		return err
	}
	if m.TelNumber, err = d.ReadString(); err != nil {
		return err
	}
	if m.AlarmRef, err = d.ReadWord16(); err != nil {
		return err
	}
	if m.AlarmSerial, err = d.ReadString(); err != nil {
		return err
	}
	if err = m.Address.unmarshal(d); err != nil {
		return err
	}
	if m.Text, err = d.ReadLongCompString(); err != nil {
		return err
	}
	return nil
}

// --- Message 40: Alert_crew ---

type AlertCrew struct {
	AlertGroup    string // 2 bytes fixed ASCII (e.g. "FA")
	ManAckReq     bool
	OpPeripherals uint16
}

func (m *AlertCrew) Type() uint8 { return MsgAlertCrew }

func (m *AlertCrew) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteFixedASCII(m.AlertGroup, 2)
	enc.WriteBool(m.ManAckReq)
	enc.WriteWord16(m.OpPeripherals)
	return enc.Bytes(), nil
}

func (m *AlertCrew) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.AlertGroup, err = d.ReadFixedASCII(2); err != nil {
		return err
	}
	if m.ManAckReq, err = d.ReadBool(); err != nil {
		return err
	}
	if m.OpPeripherals, err = d.ReadWord16(); err != nil {
		return err
	}
	return nil
}

// --- Message 42: Alert_status ---
// alerter_status can be 1-byte or 2-byte ASCII codes.

type AlertStatus struct {
	AlerterStatus string // 1 or 2 byte ASCII code
}

func (m *AlertStatus) Type() uint8 { return MsgAlertStatus }

func (m *AlertStatus) MarshalGD92() ([]byte, error) {
	return []byte(m.AlerterStatus), nil
}

func (m *AlertStatus) UnmarshalGD92(data []byte) error {
	m.AlerterStatus = string(data)
	return nil
}

// --- Message 43: Alert_eng ---

type AlertEng struct {
	AlerterEngineering byte // ASCII: A/B/Z/J/K/L/M/C/D/E
}

func (m *AlertEng) Type() uint8 { return MsgAlertEng }

func (m *AlertEng) MarshalGD92() ([]byte, error) {
	return []byte{m.AlerterEngineering}, nil
}

func (m *AlertEng) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.AlerterEngineering, err = d.ReadWord8()
	return err
}

// --- Message 50: ACK ---

type ACK struct{}

func (m *ACK) Type() uint8 { return MsgACK }

func (m *ACK) MarshalGD92() ([]byte, error) {
	return nil, nil
}

func (m *ACK) UnmarshalGD92(data []byte) error {
	return nil
}

// --- Message 51: NAK ---

type NAK struct {
	Count         uint8
	Destinations  []CommsAddress
	ReasonCodeSet uint8
	ReasonCode    uint8
}

func (m *NAK) Type() uint8 { return MsgNAK }

func (m *NAK) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.Count)
	for _, dst := range m.Destinations {
		enc.WriteCommsAddress(dst)
	}
	enc.WriteWord8(m.ReasonCodeSet)
	enc.WriteWord8(m.ReasonCode)
	return enc.Bytes(), nil
}

func (m *NAK) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Count, err = d.ReadWord8(); err != nil {
		return err
	}
	m.Destinations = make([]CommsAddress, m.Count)
	for i := range m.Destinations {
		if m.Destinations[i], err = d.ReadCommsAddress(); err != nil {
			return err
		}
	}
	if m.ReasonCodeSet, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.ReasonCode, err = d.ReadWord8(); err != nil {
		return err
	}
	return nil
}

// --- Message 60: Set_parameter ---

type SetParameter struct {
	ParameterTable uint8
	ParameterNo    uint8
	Entries        []ParameterEntry
}

func (m *SetParameter) Type() uint8 { return MsgSetParameter }

func (m *SetParameter) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.ParameterTable)
	enc.WriteWord8(m.ParameterNo)
	for _, e := range m.Entries {
		enc.WriteWord16(e.Index)
		enc.WriteBytes(e.Value)
	}
	return enc.Bytes(), nil
}

func (m *SetParameter) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.ParameterTable, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.ParameterNo, err = d.ReadWord8(); err != nil {
		return err
	}
	// Remaining data is index/value pairs - we store them as raw since
	// parameter_value type depends on the specific parameter being set
	for d.Remaining() > 0 {
		var e ParameterEntry
		if e.Index, err = d.ReadWord16(); err != nil {
			return err
		}
		// Rest of the data for this entry - since we don't know the parameter type,
		// consume remaining data as a single entry value
		if d.Remaining() > 0 {
			e.Value, err = d.ReadBytes(d.Remaining())
			if err != nil {
				return err
			}
		}
		m.Entries = append(m.Entries, e)
	}
	return nil
}

// --- Message 61: Parameter_request ---

type ParameterRequest struct {
	ParameterTable uint8
	ParameterNo    uint8
}

func (m *ParameterRequest) Type() uint8 { return MsgParameterRequest }

func (m *ParameterRequest) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.ParameterTable)
	enc.WriteWord8(m.ParameterNo)
	return enc.Bytes(), nil
}

func (m *ParameterRequest) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.ParameterTable, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.ParameterNo, err = d.ReadWord8(); err != nil {
		return err
	}
	return nil
}

// --- Message 62: Parameter ---

type Parameter struct {
	MoreValues     bool
	ParameterValue []byte // raw: type depends on the parameter
}

func (m *Parameter) Type() uint8 { return MsgParameter }

func (m *Parameter) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteBool(m.MoreValues)
	enc.WriteBytes(m.ParameterValue)
	return enc.Bytes(), nil
}

func (m *Parameter) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.MoreValues, err = d.ReadBool(); err != nil {
		return err
	}
	if d.Remaining() > 0 {
		m.ParameterValue, err = d.ReadBytes(d.Remaining())
		if err != nil {
			return err
		}
	}
	return nil
}

// --- Message 63: Param_req_multiple ---

type ParamReqMultiple struct {
	ParameterTable uint8
	ParameterNo    uint8
	FirstEntry     uint16
	LastEntry      uint16
}

func (m *ParamReqMultiple) Type() uint8 { return MsgParamReqMultiple }

func (m *ParamReqMultiple) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.ParameterTable)
	enc.WriteWord8(m.ParameterNo)
	enc.WriteWord16(m.FirstEntry)
	enc.WriteWord16(m.LastEntry)
	return enc.Bytes(), nil
}

func (m *ParamReqMultiple) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.ParameterTable, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.ParameterNo, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.FirstEntry, err = d.ReadWord16(); err != nil {
		return err
	}
	if m.LastEntry, err = d.ReadWord16(); err != nil {
		return err
	}
	return nil
}

// --- Message 64: Test ---

type Test struct {
	TestType uint8
}

func (m *Test) Type() uint8 { return MsgTest }

func (m *Test) MarshalGD92() ([]byte, error) {
	return []byte{m.TestType}, nil
}

func (m *Test) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.TestType, err = d.ReadWord8()
	return err
}

// --- Message 65: Printer_status ---

type PrinterStatusMsg struct {
	PrinterStatus uint8 // 0=offline, 1=paper_out, 2=online
}

func (m *PrinterStatusMsg) Type() uint8 { return MsgPrinterStatus }

func (m *PrinterStatusMsg) MarshalGD92() ([]byte, error) {
	return []byte{m.PrinterStatus}, nil
}

func (m *PrinterStatusMsg) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.PrinterStatus, err = d.ReadWord8()
	return err
}

// --- Message 66: MTA_status_change ---

type MTAStatusChange struct {
	MTAStatus uint8 // 0=idle, 1=online, 2=offline(user), 3=offline(fault)
}

func (m *MTAStatusChange) Type() uint8 { return MsgMTAStatusChange }

func (m *MTAStatusChange) MarshalGD92() ([]byte, error) {
	return []byte{m.MTAStatus}, nil
}

func (m *MTAStatusChange) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.MTAStatus, err = d.ReadWord8()
	return err
}

// --- Message 67: Route_status ---

type RouteStatus struct {
	Enabled          bool
	DestinationNodes []AddressRange
}

func (m *RouteStatus) Type() uint8 { return MsgRouteStatus }

func (m *RouteStatus) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteBool(m.Enabled)
	// destination_nodes: count followed by address_ranges
	enc.WriteWord8(uint8(len(m.DestinationNodes)))
	for _, ar := range m.DestinationNodes {
		enc.WriteCommsAddress(ar.Start)
		enc.WriteCommsAddress(ar.End)
	}
	return enc.Bytes(), nil
}

func (m *RouteStatus) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.Enabled, err = d.ReadBool(); err != nil {
		return err
	}
	count, err := d.ReadWord8()
	if err != nil {
		return err
	}
	m.DestinationNodes = make([]AddressRange, count)
	for i := range m.DestinationNodes {
		if m.DestinationNodes[i].Start, err = d.ReadCommsAddress(); err != nil {
			return err
		}
		if m.DestinationNodes[i].End, err = d.ReadCommsAddress(); err != nil {
			return err
		}
	}
	return nil
}

// --- Message 100: Supplier_message ---

type SupplierMessage struct {
	Data []byte // (<binary>)
}

func (m *SupplierMessage) Type() uint8 { return MsgSupplierMessage }

func (m *SupplierMessage) MarshalGD92() ([]byte, error) {
	out := make([]byte, len(m.Data))
	copy(out, m.Data)
	return out, nil
}

func (m *SupplierMessage) UnmarshalGD92(data []byte) error {
	m.Data = make([]byte, len(data))
	copy(m.Data, data)
	return nil
}

// --- Message 101: Brigade_message ---

type BrigadeMessage struct {
	Text string // long_comp_string
}

func (m *BrigadeMessage) Type() uint8 { return MsgBrigadeMessage }

func (m *BrigadeMessage) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteLongCompString(m.Text)
	return enc.Bytes(), nil
}

func (m *BrigadeMessage) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.Text, err = d.ReadLongCompString()
	return err
}

// --- Message 102: Data_base_query ---

type DataBaseQuery struct {
	QueryType byte // ASCII
	Text      string // long_comp_string
}

func (m *DataBaseQuery) Type() uint8 { return MsgDataBaseQuery }

func (m *DataBaseQuery) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.QueryType)
	enc.WriteLongCompString(m.Text)
	return enc.Bytes(), nil
}

func (m *DataBaseQuery) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.QueryType, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.Text, err = d.ReadLongCompString(); err != nil {
		return err
	}
	return nil
}

// --- Message 103: Formatted_text ---

type FormattedText struct {
	FormatType uint8
	Table      string // compressed_string, 0-255
}

func (m *FormattedText) Type() uint8 { return MsgFormattedText }

func (m *FormattedText) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.FormatType)
	enc.WriteCompressedString(m.Table)
	return enc.Bytes(), nil
}

func (m *FormattedText) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.FormatType, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.Table, err = d.ReadCompressedString(); err != nil {
		return err
	}
	return nil
}

// --- Message 104: Proforma_definition_query ---

type ProformaDefQuery struct {
	FormatType uint8
}

func (m *ProformaDefQuery) Type() uint8 { return MsgProformaDefQuery }

func (m *ProformaDefQuery) MarshalGD92() ([]byte, error) {
	return []byte{m.FormatType}, nil
}

func (m *ProformaDefQuery) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	m.FormatType, err = d.ReadWord8()
	return err
}

// --- Message 105: Proforma_definition ---

type ProformaDefinition struct {
	FormatType uint8
	Table      string // compressed_string, 0-255
}

func (m *ProformaDefinition) Type() uint8 { return MsgProformaDefinition }

func (m *ProformaDefinition) MarshalGD92() ([]byte, error) {
	enc := NewEncoder()
	enc.WriteWord8(m.FormatType)
	enc.WriteCompressedString(m.Table)
	return enc.Bytes(), nil
}

func (m *ProformaDefinition) UnmarshalGD92(data []byte) error {
	d := NewDecoder(data)
	var err error
	if m.FormatType, err = d.ReadWord8(); err != nil {
		return err
	}
	if m.Table, err = d.ReadCompressedString(); err != nil {
		return err
	}
	return nil
}
