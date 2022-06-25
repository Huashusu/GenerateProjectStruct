package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
)

type Dir struct {
	Name  string
	Perm  fs.FileMode
	GoMod bool
}

type DirConfig struct {
	ProjectName string
	Root        string
	Perm        fs.FileMode
	Dirs        []Dir
}

var GoRootEnv = "GOROOT=" + runtime.GOROOT()

// GenerateDirList 用于创建项目文件夹结构
func GenerateDirList(c DirConfig) {
	projectPath := path.Join(c.Root, c.ProjectName)
	err := os.MkdirAll(projectPath, c.Perm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	msg, err := initGoMod(c.ProjectName)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(msg)
	log.Println("Generate project root directory success")
	maxLength := 0
	for _, dir := range c.Dirs {
		length := len(path.Join(projectPath, dir.Name))
		if maxLength < length {
			maxLength = length
		}
	}
	formatStr := fmt.Sprintf("Generate directory %%-%ds success\n", maxLength+1)
	for _, dir := range c.Dirs {
		dirPath := path.Join(projectPath, dir.Name)
		err := os.MkdirAll(dirPath, dir.Perm)
		if err != nil && !os.IsExist(err) {
			log.Printf("Generate directory error:%+v\n", err)
		}
		if dir.GoMod {
			msg, err := initGoMod(dirPath)
			if err != nil {
				log.Printf("Directory go mod init error:%+v\n", err)
			}
			log.Print(msg)
		}
		log.Printf(formatStr, dirPath)
	}
	return
}

func initGoMod(path string) (string, error) {
	cmd := exec.Command("go", "mod", "init", path)
	cmd.Dir = path
	cmd.Env = []string{GoRootEnv}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait()
	return out.String(), err
}
