package plumbing

import (
	"log"
	"os"
	"crypto/sha1"
    "encoding/hex"
	"fmt"
)

func readFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileSize := len(content)
	header := fmt.Sprintf("blob %d\x00", fileSize)
	byteStream := append([]byte(header), content...)
	return byteStream, nil
}

func HashFile(filePath string) {
	stream, err := readFile(filePath)
	if err != nil {
		log.Fatalf("Error while getting content of the %s: %s", filePath, err)
	}

	hash := sha1.New()
	hash.Write(stream)
	sha1_hash := hex.EncodeToString(hash.Sum(nil))
	fmt.Println("Hash: ", sha1_hash)
}
