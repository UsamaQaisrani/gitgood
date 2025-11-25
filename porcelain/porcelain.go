package porcelain	

import (
	"log"
	"os"
	"fmt"
	"usamaqaisrani/git-good/plumbing"
)

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

func createDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		log.Fatal("Error occured while create refs directory:", err)
		return
	}
}

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

func writeObject(dirName, fileName string, content []byte) {
	filePath := objects + "/" + dirName + "/" + fileName
	fmt.Println(filePath)
	createDir(objects + "/" + dirName)
	writeFile(filePath, content)
}

func Stage(filePath string) {
	stream, err := plumbing.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error while getting content of the %s: %s", filePath, err)
	}

	hash := plumbing.HashFile(stream)
	if err != nil {
		log.Fatalf("Error while hashing %s: %s", filePath, err)
	}

	compressed, err := plumbing.Compress(stream) 
	if err != nil {
		log.Fatalf("Error while compressing %s: %s", filePath, err)
	}
	fmt.Printf("Compressed (%d bytes): % x\n", len(compressed), compressed)

	dirName := hash[:2]
	fileName := hash[3:]
	writeObject(dirName, fileName, compressed)
}
