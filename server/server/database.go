package server

import (
	"fmt"
	"log"
)

type rawPulse []int
type rawPulses []rawPulse
type frameList []rawPulses
type valuePulseListMap map[string]frameList
type protocolValueMap map[protocolID]valuePulseListMap
type frameDatabase map[string]protocolValueMap

type frameCRUD interface {
	insert(pTaggedFrame taggedFrame) error
	getCollectorIDList() ([]string, error)
	getProtocolIDList(pCollectorID string) ([]protocolID, error)
	getValues(pCollectorID string, pProtocolID protocolID) ([]string, error)
	getFrameList(pCollectorID string, pProtocolID protocolID, value string) (frameList, error)
}

func newDatabase() frameCRUD {
	db := make(frameDatabase)
	return &db
}

func (db *frameDatabase) insert(pTaggedFrame taggedFrame) error {
	protocol, value, err := decodeFrame(&pTaggedFrame)
	if err != nil {
		return err
	}
	dbase := *db
	protocol2Value, collectorIDOk := dbase[pTaggedFrame.CollectorID]
	if !collectorIDOk {
		log.Printf("Collector ID '%s' not found. Creating new entry.", pTaggedFrame.CollectorID)
		dbase[pTaggedFrame.CollectorID] = make(protocolValueMap)
		protocol2Value = dbase[pTaggedFrame.CollectorID]
	}
	value2FrameList, protocolIDOk := protocol2Value[protocol]
	if !protocolIDOk {
		log.Printf("Protocol ID '%s' not found. Creating new entry.", protocol)
		protocol2Value[protocol] = make(valuePulseListMap)
		value2FrameList = protocol2Value[protocol]
	}
	frames, valueOk := value2FrameList[value]
	if !valueOk {
		log.Printf("Frame value '%s' not found. Creating new entry.", value)
		value2FrameList[value] = make(frameList, 0, 1)
		frames = value2FrameList[value]
	}
	log.Printf("Inserting %+v.", pTaggedFrame.Frame.Data)
	pulses := make(rawPulses, len(pTaggedFrame.Frame.Data))
	for i, d := range pTaggedFrame.Frame.Data {
		pulses[i] = make(rawPulse, 2)
		pulses[i][0] = d[0] * pTaggedFrame.Frame.Resolution
		pulses[i][1] = d[1] * pTaggedFrame.Frame.Resolution
	}
	frames = append(frames, pulses)
	value2FrameList[value] = frames
	log.Printf("%s > %s > %s now has %d items", pTaggedFrame.CollectorID, protocol, value, len(value2FrameList[value]))
	return nil
}

func (db *frameDatabase) getCollectorIDList() ([]string, error) {
	dbase := *db
	collectorIDList := make([]string, 0, len(dbase))
	for cid := range dbase {
		collectorIDList = append(collectorIDList, cid)
	}
	return collectorIDList, nil
}

func (db *frameDatabase) getProtocolIDList(pCollectorID string) ([]protocolID, error) {
	dbase := *db
	protocol2Value, collectorIDOk := dbase[pCollectorID]
	if !collectorIDOk {
		return nil, fmt.Errorf("Collector ID '%s' cannot be found", pCollectorID)
	}
	protocolIDList := make([]protocolID, 0, len(protocol2Value))
	for pid := range protocol2Value {
		protocolIDList = append(protocolIDList, pid)
	}
	return protocolIDList, nil
}

func (db *frameDatabase) getValues(pCollectorID string, pProtocolID protocolID) ([]string, error) {
	dbase := *db
	protocol2Value, collectorIDOk := dbase[pCollectorID]
	if !collectorIDOk {
		return nil, fmt.Errorf("Collector ID '%s' cannot be found", pCollectorID)
	}
	value2FrameList, protocolIDOk := protocol2Value[pProtocolID]
	if !protocolIDOk {
		return nil, fmt.Errorf("Protocol ID '%s' cannot be found", pProtocolID)
	}
	valueList := make([]string, 0, len(value2FrameList))
	for value := range value2FrameList {
		valueList = append(valueList, value)
	}
	return valueList, nil
}

func (db *frameDatabase) getFrameList(pCollectorID string, pProtocolID protocolID, value string) (frameList, error) {
	dbase := *db
	protocol2Value, collectorIDOk := dbase[pCollectorID]
	if !collectorIDOk {
		return nil, fmt.Errorf("Collector ID '%s' cannot be found", pCollectorID)
	}
	value2FrameList, protocolIDOk := protocol2Value[pProtocolID]
	if !protocolIDOk {
		return nil, fmt.Errorf("Protocol ID '%s' cannot be found", pProtocolID)
	}
	frames, valueOk := value2FrameList[value]
	if !valueOk {
		return nil, fmt.Errorf("Value '%s' cannot be found", value)
	}
	return frames, nil
}
