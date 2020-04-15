package server

type frameStore struct {
	frameSet map[frameID][][]markSpacePair
}

type frameID struct {
	protocolName string
	value        string
}

func (f *frameStore) add(protocolName string, value string, header markSpacePair, mspair []markSpacePair) {
	fid := frameID{protocolName, value}

	if mspList, ok := f.frameSet[fid]; !ok {
		mspList = make([][]markSpacePair, 0)
		f.frameSet[fid] = mspList
	}

	mspairCopy := make([]markSpacePair, len(mspair))
	copy(mspairCopy, mspair)

	mspairWithHeader := make([]markSpacePair, 0)
	mspairWithHeader = append(mspairWithHeader, header)
	mspairWithHeader = append(mspairWithHeader, mspairCopy...)

	f.frameSet[fid] = append(f.frameSet[fid], mspairWithHeader)
}
