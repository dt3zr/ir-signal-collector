package server

// type frame struct {
// 	Head      markSpacePair   `json:"head"`
// 	RawPulses []markSpacePair `json:"raw-pulses"`
// }

// type frameMetadata struct {
// 	ProtocolName string `json:"protocol-name"`
// 	Value        string `json:"value"`
// 	Length       int    `json:"length"`
// }

// type signalQueryResponse struct {
// 	ProtocolName string  `json:"protocol-name"`
// 	Value        string  `json:"value"`
// 	Frames       []frame `json:"frames"`
// }

// type signalListResponse struct {
// 	Metadata []frameMetadata `json:"metadata"`
// }

type simpleValueLength struct {
	Value  string `json:"value"`
	Length int    `json:"length"`
}
type simpleValueLengthList []simpleValueLength
type simpleProtocolValueMap map[string]simpleValueLengthList
type simpleCollectorProtocolMap map[string]simpleProtocolValueMap
