package plumbing

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Paths
const git = ".gitgood"
const config = git + "/config"
const refs = git + "/refs"
const refHeads = refs + "/heads"
const head = git + "/HEAD"
const objects = git + "/objects"
const index = git + "/index"

type StageEntry struct {
	CTimeSec  uint32
	CTimeNano uint32
	MTimeSec  uint32
	MTimeNano uint32
	Dev       uint32
	Ino       uint32
	Mode      uint32
	Uid       uint32
	Gid       uint32
	Size      uint32
	Hash      string
	Flags     uint16
	Path      string
}

type FileEntry struct {
	Path    string
	Content []byte
	Info    fs.FileInfo
	Err     error
}

func ReadFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func appendHeader(stream []byte) []byte {
	fileSize := len(stream)
	header := fmt.Sprintf("blob %d\x00", fileSize)
	byteStream := append([]byte(header), stream...)
	return byteStream
}

// Create a directory at given path if it doesn't exist already
func CreateDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatal("Error occured while create refs directory:", err)
		return
	}
}

// Write a file at given path with the content
func WriteFile(path string, content any) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatal("Error occured while creating HEAD file:", err)
		return
	}
	defer f.Close()

	switch v := content.(type) {
	case string:
		_, err = f.WriteString(v)
	case []byte:
		_, err = f.Write(v)
	default:
		fmt.Println("Unrecognized content type while writing file.")
	}
	if err != nil {
		log.Fatalf("Error occured while writing %s: %s", path, err)
		return
	}
}

func HashFile(content []byte) string {
	hash := sha1.New()
	hash.Write(content)
	sha1_hash := hex.EncodeToString(hash.Sum(nil))
	fmt.Println("Hash: ", sha1_hash)
	return sha1_hash
}

func Compress(data []byte) ([]byte, error) {
	var buff bytes.Buffer
	w := zlib.NewWriter(&buff)

	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func CreateIndexInstance(path, hash string) (StageEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return StageEntry{}, err
	}

	sec := uint32(info.ModTime().Unix())
	nano := uint32(info.ModTime().Nanosecond())

	return StageEntry{
		CTimeSec:  sec,
		CTimeNano: nano,
		MTimeSec:  sec,
		MTimeNano: nano,
		Dev:       0,
		Ino:       0,
		Mode:      0x81A4,
		Uid:       0,
		Gid:       0,
		Size:      uint32(info.Size()),
		Hash:      hash,
		Path:      path,
	}, nil
}

func CreateHeaderForIndex(count int) []byte {
	header := make([]byte, 12)
	copy(header[0:4], []byte("DIRC"))
	binary.BigEndian.PutUint32(header[4:8], 2)
	binary.BigEndian.PutUint32(header[8:12], uint32(count))
	return header
}

func CreateStagingEntry(entry StageEntry) []byte {
	var buffer bytes.Buffer

	binary.Write(&buffer, binary.BigEndian, entry.CTimeSec)
	binary.Write(&buffer, binary.BigEndian, entry.CTimeNano)
	binary.Write(&buffer, binary.BigEndian, entry.MTimeSec)
	binary.Write(&buffer, binary.BigEndian, entry.MTimeNano)
	binary.Write(&buffer, binary.BigEndian, entry.Dev)
	binary.Write(&buffer, binary.BigEndian, entry.Ino)
	binary.Write(&buffer, binary.BigEndian, entry.Mode)
	binary.Write(&buffer, binary.BigEndian, entry.Uid)
	binary.Write(&buffer, binary.BigEndian, entry.Gid)
	binary.Write(&buffer, binary.BigEndian, entry.Size)

	hashBytes, _ := hex.DecodeString(entry.Hash)
	buffer.Write(hashBytes)

	pathLen := len(entry.Path)
	if pathLen > 0xFFF {
		pathLen = 0xFFF
	}

	binary.Write(&buffer, binary.BigEndian, uint16(pathLen))
	buffer.WriteString(entry.Path)

	buffer.WriteByte(0x00)
	totalLen := 62 + pathLen + 1
	padding := (8 - (totalLen % 8)) % 8

	buffer.Write(make([]byte, padding))
	return buffer.Bytes()
}

func UpdateIndex(entries []StageEntry) error {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})

	var indexBuf bytes.Buffer
	indexBuf.Write(CreateHeaderForIndex(len(entries)))
	for _, entry := range entries {
		indexBuf.Write(CreateStagingEntry(entry))
	}
	digest := sha1.Sum(indexBuf.Bytes())
	indexBuf.Write(digest[:])
	if _, err := os.Stat(git); os.IsNotExist(err) {
		os.Mkdir(git, 0755)
	}

	return os.WriteFile(index, indexBuf.Bytes(), 0644)
}

func WriteBlob(content []byte, hash string) error {
	compressed, err := Compress(content)
	if err != nil {
		return err
	}
	fmt.Printf("Compressed (%d bytes): % x\n", len(compressed), compressed)

	dirName := hash[:2]
	fileName := hash[2:]
	fullBlobPath := filepath.Join(objects, dirName, fileName)
	CreateDir(filepath.Join(objects, dirName))
	WriteFile(fullBlobPath, compressed)

	return nil
}

func WalkDir(rootPath string) <-chan FileEntry {
	entries := make(chan FileEntry)
	go func() {
		defer close(entries)

		_ = filepath.WalkDir(rootPath, func(currPath string, d fs.DirEntry, err error) error {
			if err != nil {
				entries <- FileEntry{Err: err}
				return nil
			}

			if d.IsDir() {
				if d.Name() == ".git" || d.Name() == git {
					return filepath.SkipDir
				}
				return nil
			}

			normPath := filepath.ToSlash(currPath)
			if strings.Contains(normPath, ".git") || strings.Contains(normPath, git) {
				return nil
			}

			content, err := ReadFile(currPath)
			content = appendHeader(content)
			if err != nil {
				entries <- FileEntry{Err: err}
				return nil
			}

			info, err := d.Info()
			if err != nil {
				entries <- FileEntry{Err: err}
				return nil
			}

			fileEntry := FileEntry{
				Path:    currPath,
				Content: content,
				Info:    info,
				Err:     nil,
			}
			entries <- fileEntry
			return nil
		})
	}()
	return entries
}
