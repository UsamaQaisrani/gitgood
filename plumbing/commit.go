package plumbing

import (
	"encoding/binary"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"path/filepath"
	"strings"
)

type Node struct {
	Name     string
	Mode     uint32
	Hash     string
	Path 	 string
	Children []*Node
}

func ReadIndex() error {
	content, err := ReadFile(index)
	if err != nil {
		return err
	}

	header := string(content[0:4])
	if string(header) != "DIRC" {
		return errors.New("Invalid header, not an index file.")
	}

	entryCount := binary.BigEndian.Uint32(content[8:12])

	var entries []StageEntry
	i := 12

	for j := 0; j < int(entryCount); j++ {
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
		hash := hex.EncodeToString(content[i : i+20])
		i += 20
		pathLen := binary.BigEndian.Uint16(content[i : i+2])
		i += 2
		path := string(content[i : i+int(pathLen)])

		i += int(pathLen)

		// Skipping the null byte (0x00)
		i += 1

		// Reading the padding
		entrySize := 62 + int(pathLen) + 1
		padding := (8 - (entrySize % 8)) % 8
		i += padding

		fmt.Printf("%d %s %d %s\n", mode, hash, 0, path)

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

func CreateDirTree() (*Node, error) {
	root := "."
	tree := &Node {
		Name: filepath.Base(root),
		Mode: 0x81A4,
	}

	nodeMap := map[string]*Node{
		root:tree,
	}

	err := filepath.WalkDir(root, func(currPath string, d fs.DirEntry, err error) error {
			normPath := filepath.ToSlash(currPath)
			if strings.Contains(normPath, ".git") || strings.Contains(normPath, git) {
				return nil
			}

			parent := nodeMap[filepath.Dir(currPath)]
			node := &Node{
				Name: d.Name(),
				Mode: 0x81A4,
			}

			if !d.IsDir() {
				content, err := ReadFile(currPath)
				if err != nil {
					return err
				}
				hash := HashFile(content)
				node.Hash = hash
				node.Path = currPath
			}
			
			parent.Children = append(parent.Children, node)

			if d.IsDir() {
				nodeMap[currPath] = node
			}

		return nil
	})

	if err != nil {
		return tree, err
	}

	return tree, nil
}

func BuildObject(node *Node) (string, error) {

	//If its a file with no children
	if len(node.Children) < 1 {
		return node.Hash, nil
	}

	var stageEntries []StageEntry

	for _, child := range node.Children {
		hash, err := BuildObject(child)
		if err != nil {
			return "", err
		}

		entry := StageEntry {
			Name: child.Name,
			Mode: child.Mode,
			Hash: hash,
		}
		stageEntries = append(stageEntries,entry)
	}

	sort.Slice(stageEntries, func(i, j int) bool {
		return stageEntries[i].Name < stageEntries[j].Name
	})

	var buff bytes.Buffer

	for _, child := range stageEntries {
		modeString := fmt.Sprintf("%o", child.Mode)
		buff.WriteString(modeString) //Writing mode as string

		buff.WriteByte(0x20) //Writing space as byte
		buff.WriteString(child.Name)
		buff.WriteByte(0x00) //Writing null byte

		hashBytes, _ := hex.DecodeString(child.Hash)
		buff.Write(hashBytes)
	}

	buffSize := buff.Len()
	header := fmt.Sprintf("tree %d\x00", buffSize)
	fullData := append([]byte(header), buff.Bytes()...)

	treeHash := HashFile(fullData)

	err := WriteBlob(fullData, treeHash)
	if err != nil {
		return "", nil
	}
	return treeHash, nil
}

