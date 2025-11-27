package plumbing

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	i := 12

	//TODO:- Iterate entry count times on rest of the data to get entries
	for j:=0; j<7 ; j++{
		cTimeSec := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		cTimeNano := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		mTimeSec := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		mTimeNano := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		dev := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		ino := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		mode := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		uid := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		gid := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		size := binary.BigEndian.Uint32(content[i : i+4])
		i += 4
		hash := string(content[i : i+20])
		i += 20
		pathLen := binary.BigEndian.Uint16(content[i:i+2])
		i += 2
		path := string(content[i:i+int(pathLen)])

		i += int(pathLen)

		// Reading the null byte
		i += 1

		// Reading the padding
		entrySize := 62 + int(pathLen) + 1
		padding := (8 - (entrySize % 8)) % 8
		i += padding

		fmt.Println(path)

		entry := StageEntry{
			CTimeSec:  cTimeSec,
			CTimeNano: cTimeNano,
			MTimeSec:  mTimeSec,
			MTimeNano: mTimeNano,
			Dev:       dev,
			Ino:       ino,
			Mode:      mode,
			Uid:       uid,
			Gid:       gid,
			Size:      size,
			Hash:      hash,
			Path:      path,
		}

		entries = append(entries, entry)
	}

	return nil
}
