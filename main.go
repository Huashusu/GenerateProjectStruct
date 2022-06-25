package main

import (
	"fmt"
	"log"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	fmt.Printf("Project name: ")
	var name string
	_, err := fmt.Scanf("%s\n", &name)
	if err != nil {
		return
	}
	config := DirConfig{
		ProjectName: name,
		Root:        "./",
		Perm:        0744,
		Dirs: []Dir{
			{"cmd", 0744, true},
			{"api", 0744, false},
			{"core", 0744, false},
			{"config", 0744, false},
			{"global", 0744, false},
			{"initialiaze", 0744, false},
			{"log", 0744, false},
			{"middleware", 0744, false},
			{"model", 0744, false},
			{"router", 0744, false},
			{"service", 0744, false},
			{"utils", 0744, false},
		},
	}
	GenerateDirList(config)
	fmt.Printf("Generate default code (y/n): ")
	var generateCode string
	_, err = fmt.Scanf("%s\n", &generateCode)
	if err != nil {
		return
	}
	if generateCode == "y" {
		GenerateCode(name)
	}
}
