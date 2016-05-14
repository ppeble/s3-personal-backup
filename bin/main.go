package main

import (
	"flag"
	"os"

	"github.com/ptrimble/dreamhost-personal-backup"
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

	localFileProcessor := backup.NewLocalFileProcessor()
	// Create Remote processor

	processor := backup.NewProcessor(localFileProcessor.Process)

	processor.Process(targetDir)

	// a processor should do this
	// Get all local info. - DONE (I think)
	//Go through EVERY file and compare it against the remote. This can be a comparer that decides what to do
	// a) If diff found (or not found) then add it to the 'diff' list
	// b) if no diff then move on
	// c) At the end, send everything in the 'diff' list
	// Keep track of what changed to give a report at the end
}
