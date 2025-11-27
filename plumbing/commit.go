package plumbing

import (
	"fmt"
	"errors"
	"encoding/binary"
)

func ReadIndex() error {
	content, err := ReadFile(index)
	if err != nil {
		return err
	}

	header := string(content[0:4])
	if string(header) != "DIRC" {
		return errors.New("Invalid header, not an index file.")
	}

	signature := binary.BigEndian.Uint32(content[4:8])
	entryCount := binary.BigEndian.Uint32(content[8:12])
	fmt.Println(header)
	fmt.Println("Version:", signature)
	fmt.Println("Entires count:", entryCount)

	var entries []StageEntry

	//TODO:- Iterate entry count times on rest of the data to get entries
	for i:=12; i<int(entryCount)*4; i++ {
		cTimeSec := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		cTimeNano := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		mTimeSec := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		mTimeNano := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		dev := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		ino := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		mode := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		uid := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		gid := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		size := binary.BigEndian.Uint32(content[i:i+4])
		i += 4
		hash := string(content[i:i+4])
		i += 4
		path := string(content[i:i+4])
		i += 4

		entry := StageEntry {
			CTimeSec: cTimeSec,
			CTimeNano: cTimeNano,
			MTimeSec: mTimeSec,
			MTimeNano: mTimeNano,
			Dev: dev,
			Ino: ino,
			Mode: mode,
			Uid: uid,
			Gid: gid,
			Size: size,
			Hash: hash,
			Path: path,
		}

		entries = append(entries, entry)
	}
	fmt.Println(len(entries))

	return nil
}
