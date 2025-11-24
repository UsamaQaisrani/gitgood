package porcelain	

import (
	"log"
	"os"
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

func writeFile(path, content string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Error occured while creating HEAD file:", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		log.Fatalf("Error occured while writing %s: %s", path, err)
		return
	}
}

func Stage(filePath string) {
	plumbing.HashFile(filePath)
}
