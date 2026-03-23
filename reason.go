package gd92

// ReasonCodeSet identifies which category of reason codes applies (spec Appendix D).
const (
	ReasonSetPrinter    = 0
	ReasonSetGeneral    = 1
	ReasonSetAlerter    = 2
	ReasonSetPeripheral = 3
	ReasonSetParameter  = 4
)

// Alerter reason codes (D.1).
const (
	AlerterTxAFail       = 1
	AlerterTxBFail       = 2
	AlerterTxFail        = 3
	AlerterGrAFail       = 4
	AlerterGrBFail       = 5
	AlerterGrCFail       = 6
	AlerterEncFail       = 7
	AlerterPowerFail     = 8
	AlerterOtherFail     = 9
	AlerterSysBusy       = 10
	AlerterNoFac         = 11
	AlerterOtherProb     = 12
	AlerterWaitAck       = 13
	AlerterNotConnected  = 14
)

// General reason codes (D.2).
const (
	GeneralNoBearer     = 1
	GeneralInvParam     = 2
	GeneralInvMess      = 3
	GeneralAvlFail      = 4
	GeneralStatusFail   = 5
	GeneralTextType     = 6
	GeneralInvCall      = 7
	GeneralNotAccepted  = 8
	GeneralWaitAck      = 9
	GeneralTestFail     = 10
	GeneralInvTest      = 11
	GeneralInvProt      = 12
	GeneralCheckError   = 13
	GeneralAbort        = 14
	GeneralNoPort       = 15
	GeneralDataFail     = 16
	GeneralProformFail  = 17
)

// Peripheral reason codes (D.3).
const (
	PeripheralNoDoor         = 1
	PeripheralNoSounder      = 2
	PeripheralNoLights       = 3
	PeripheralNoInd1         = 4
	PeripheralNoInd2         = 5
	PeripheralNoInd3         = 6
	PeripheralNoInd4         = 7
	PeripheralNoInd5         = 8
	PeripheralNoInd6         = 9
	PeripheralNoInd7         = 10
	PeripheralNoInd8         = 11
	PeripheralDoorFail       = 12
	PeripheralSounderFail    = 13
	PeripheralLightsFail     = 14
	PeripheralInd1Fail       = 15
	PeripheralInd2Fail       = 16
	PeripheralInd3Fail       = 17
	PeripheralInd4Fail       = 18
	PeripheralInd5Fail       = 19
	PeripheralInd6Fail       = 20
	PeripheralInd7Fail       = 21
	PeripheralInd8Fail       = 22
	PeripheralOperationFail  = 23
	PeripheralNoDevResponse  = 24
)

// Printer reason codes (D.4).
const (
	PrinterNoPrintRes    = 1
	PrinterOffLine       = 2
	PrinterNoPaper       = 3
	PrinterLowPaper      = 4
	PrinterNoPower       = 5
	PrinterNoConnection  = 6
)

// Parameter reason codes (D.5).
const (
	ParamNoModAccess     = 1
	ParamInvalidSyntax   = 2
	ParamInvalidValue    = 3
	ParamInvalidPassword = 4
	ParamInvalidTable    = 5
	ParamInvalidParam    = 6
)
