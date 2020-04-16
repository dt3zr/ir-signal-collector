package collector

// signalPublishRequest contains the description of a frame of
// infrared signal to be published to the server
type signalPublishRequest struct {
	ProtocolName string          `json:"protocol-name"`
	FrameSize    int             `json:"frame-size"`
	Value        string          `json:"value"`
	Header       markSpacePair   `json:"header"`
	RawPulses    []markSpacePair `json:"raw-pulses"`
}

// markSpacePair represents each bit for the frame
type markSpacePair struct {
	Mark  float64 `json:"mark"`
	Space float64 `json:"space"`
}
