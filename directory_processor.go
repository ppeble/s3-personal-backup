package main

import (
	"io/ioutil"
	"os"
)

func processDirectories(dirChan chan []os.FileInfo, targetDir string) {
	targetDirContents, _ := ioutil.ReadDir(targetDir)

	dirChan <- targetDirContents
}
