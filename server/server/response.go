package server

//
type simpleValueLength struct {
	Value  string `json:"value"`
	Length int    `json:"length"`
}
type simpleValueLengthList []simpleValueLength
type simpleProtocolValueMap map[string]simpleValueLengthList
type simpleCollectorProtocolMap map[string]simpleProtocolValueMap

type newFrameEvent struct {
	CollectorID string    `json:"collectorId"`
	ProtocolID  string    `json:"protocolID"`
	Value       string    `json:"value"`
	Frame       rawPulses `json:"frame"`
}
