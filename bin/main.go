package main

import (
	"flag"
	"os"

	"github.com/minio/minio-go"

	backup "github.com/ptrimble/dreamhost-personal-backup"
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

	//TODO Set up these values are args
	s3Client, err := minio.NewV2("objects-us-west-1.dream.io", "hXQheR4_EeBgkX7GgINx", "4kXhKcPmIRSXAXR_DSJwhCFQkc2X49N6q5SHvkGv", false)
	if err != nil {
		panic(err)
	}

	localFileProcessor := backup.NewLocalFileProcessor(targetDir)

	//TODO The bucket needs to be A) an arg or B) picked at random (uuid?)
	remoteFileProcessor, err := backup.NewRemoteFileProcessor(
		"test11112",
		s3Client.ListObjects,
		s3Client.RemoveObject,
		s3Client.FPutObject,
	)

	if err != nil {
		panic(err)
	}

	processor := backup.NewProcessor(
		localFileProcessor.Gather,
		remoteFileProcessor.Gather,
	)

	processor.Process()

	// a processor should do this
	// Get all local info. - DONE
	// Get all remote info - DONE
	//Go through EVERY file and compare it against the remote. This can be a comparer that decides what to do
	// a) If diff found (or not found) then add it to the 'diff' list
	// b) if no diff then move on
	// c) At the end, send everything in the 'diff' list
	// Keep track of what changed to give a report at the end
}
