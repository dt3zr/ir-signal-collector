package server

//
type simpleValueLength struct {
	Value  string `json:"value"`
	Length int    `json:"length"`
}
type simpleValueLengthList []simpleValueLength
type simpleProtocolValueMap map[string]simpleValueLengthList
type simpleCollectorProtocolMap map[string]simpleProtocolValueMap

type value2FrameListMap map[string]frameList
type protocol2ValueMap map[string]value2FrameListMap
type collector2ProtocolMap map[string]protocol2ValueMap

type newFrameEvent struct {
	CollectorID string `json:"collectorId"`
	ProtocolID  string `json:"protocolID"`
	Value       string `json:"value"`
	Frame       frame  `json:"frame"`
}
