package server

import (
	"errors"
	"fmt"
	"math"
)

type protocolID int

const (
	protocolNEC protocolID = iota
	protocolUnknown
)

func (pid protocolID) String() string {
	switch pid {
	case protocolNEC:
		return "NEC"
	default:
		return "Unknown"
	}
}

func (pid *protocolID) parse(protocolIDString string) {
	switch protocolIDString {
	case "NEC":
		*pid = protocolNEC
	default:
		*pid = protocolUnknown
	}
}

func decodeFrame(pTaggedFrame *taggedFrame) (protocolID, string, error) {
	var err error
	pid := protocolUnknown
	decodedValue := ""
	if header, err := pTaggedFrame.getHeader(); err == nil {
		if matchNECProtocol(header) {
			pid = protocolNEC
			decodedValue, err = decodedNECFrameValue(pTaggedFrame.getRawPulses())
		}
	}
	return pid, decodedValue, err
}

const (
	pNECHeaderMarkMicros float64 = 9000
	pNECBitShortMicros   float64 = 562.5
	pNECBitLongMicros    float64 = 1687.5
)

func matchNECProtocol(h *markSpacePair) bool {
	return (float64(h.Mark) > pNECHeaderMarkMicros*0.90) && (float64(h.Mark) < pNECHeaderMarkMicros*1.1)
}

func matchNECShort(microTime float64) bool {
	return (float64(microTime) > pNECBitShortMicros*0.90) && (float64(microTime) < pNECBitShortMicros*1.1)
}

func matchNECLong(microTime float64) bool {
	return (float64(microTime) > pNECBitLongMicros*0.90) && (float64(microTime) < pNECBitLongMicros*1.1)
}

func decodedNECFrameValue(rawPulses []markSpacePair) (string, error) {
	var decodedValue uint32 = 0
	if len(rawPulses[1:]) < 32 {
		return "", errors.New("NEC has less than 32 raw pulses")
	}
	pulses := rawPulses[1:33]
	for i, p := range pulses {
		isMarkShort := matchNECShort(p.Mark)
		isSpaceShort := matchNECShort(p.Space)
		isSpaceLong := matchNECLong(p.Space)
		if isMarkShort && isSpaceLong {
			decodedValue = decodedValue + uint32(math.Pow(float64(2), float64(31-i)))
		} else if isMarkShort && !isSpaceShort {
			return "", errors.New("Error decoding NEC frame value")
		}
	}
	return fmt.Sprintf("%08X", decodedValue), nil

}
