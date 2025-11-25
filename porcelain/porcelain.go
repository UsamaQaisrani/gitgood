package porcelain	

import (
	"log"
	"os"
	"fmt"
	"path"
	"usamaqaisrani/git-good/plumbing"
)

// Paths
const git = ".gitgood"
const config = git + "/config"
const refs = git + "/refs"
const refHeads = refs + "/heads"
const head = git + "/HEAD"
const objects = git + "/objects"

func Init() {
	createDir(git)
	createDir(refs)
	createDir(refHeads)
	createDir(objects)
	headContent := "ref: refs/heads/master\n"
	configContent := "[core]\n\trepositoryformatversion = 0\n\tfilemode = true\n\tbare = false\n\tlogallrefupdates = true"
	writeFile(head, headContent)
	writeFile(config, configContent)
}

//Create a directory at given path if it doesn't exist already
func createDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		log.Fatal("Error occured while create refs directory:", err)
		return
	}
}

// Write a file at given path with the content
func writeFile(path string, content any) {
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

// Stage the file at given path
func Stage(path string) {
	stream, err := plumbing.ReadFile(path)
	if err != nil {
		log.Fatalf("Error while getting content of the %s: %s", path, err)
		return
	}

	hash := plumbing.HashFile(stream)
	if err != nil {
		log.Fatalf("Error while hashing %s: %s", path, err)
		return
	}

	compressed, err := plumbing.Compress(stream) 
	if err != nil {
		log.Fatalf("Error while compressing %s: %s", path, err)
		return
	}
	fmt.Printf("Compressed (%d bytes): % x\n", len(compressed), compressed)

	dirName := hash[:2]
	fileName := hash[3:]
	fullBlobPath := path.Join(objects, dirName, fileName)
	createDir(path.Join(objects, dirName))
	writeFile(fullBlobPath, compressed)
}
