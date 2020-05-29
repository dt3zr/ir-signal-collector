package server

import (
	"errors"
	"fmt"
)

type taggedFrame struct {
	CollectorID string    `json:"collectorId"`
	Frame       frameData `json:"frame"`
}

type frameData struct {
	Resolution int     `json:"resolution"`
	Data       [][]int `json:"data"`
}

// markSpacePair represents each bit for the frame
type markSpacePair struct {
	Mark  float64 `json:"mark"`
	Space float64 `json:"space"`
}

func (m markSpacePair) String() string {
	return fmt.Sprintf("(%v, %v)", m.Mark, m.Space)
}

func (f *taggedFrame) getHeader() (*markSpacePair, error) {
	if len(f.Frame.Data) < 1 {
		return nil, errors.New("Frame has not header")
	}

	return &markSpacePair{
		float64(f.Frame.Data[0][0] * f.Frame.Resolution),
		float64(f.Frame.Data[0][1] * f.Frame.Resolution),
	}, nil
}

func (f *taggedFrame) getRawPulses() []markSpacePair {
	pulses := f.Frame.Data
	pair := make([]markSpacePair, len(pulses))

	for i, p := range pulses {
		pair[i].Mark = float64(p[0] * f.Frame.Resolution)
		pair[i].Space = float64(p[1] * f.Frame.Resolution)
	}

	return pair
}
