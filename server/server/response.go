package server

type frame struct {
	Head      markSpacePair   `json:"head"`
	RawPulses []markSpacePair `json:"raw-pulses"`
}

type signalQueryResponse struct {
	ProtocolName string  `json:"protocol-name"`
	Value        string  `json:"value"`
	Frames       []frame `json:"frames"`
}
