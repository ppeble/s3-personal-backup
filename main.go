package main

import (
	"flag"
	//"fmt"
	//"io/ioutil"
	"os"
)

func main() {
	var targetDirViaFlag, targetDirViaEnv string
	flag.StringVar(&targetDirViaFlag, "targetDir", "", "Local directory to back up. Required.")
	flag.Parse()

	targetDirViaEnv = os.Getenv("PERSONAL_BACKUP_TARGET_DIR")

	var targetDir string
	if targetDirViaFlag != "" {
		targetDir = targetDirViaFlag
	} else if targetDirViaEnv != "" {
		targetDir = targetDirViaEnv
	} else {
		panic("target dir must be specified via either command line (-targetDir) or env var (PERSONAL_BACKUP_TARGET_DIR)")
	}

	dirChannel := make(chan []os.FileInfo)
	go processDirectories(dirChannel, targetDir)
}
