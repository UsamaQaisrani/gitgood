package main

import (
	"fmt"
	"os"
	"usamaqaisrani/git-good/porcelain"
)

var commands = map[string]struct{}{
    "init": {},
}

func main() {
    args := os.Args[1:]

    if len(args) == 0 {
        fmt.Println("No command provided.")
        return
    }

    if _, ok := commands[args[0]]; !ok {
        fmt.Printf("%s command not defined\n", args[0])
        return
    }

	if args[0] == "init" {
		porcelain.Init()
	}
}
