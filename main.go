package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/minio/minio-go"

	"github.com/ptrimble/dreamhost-personal-backup/backup"
	"github.com/ptrimble/dreamhost-personal-backup/backup/worker"
)

var targetDir string

var s3Host, s3AccessKey, s3SecretKey, s3BucketName string

func main() {
	processVars()

	s3Client, err := minio.NewV2(s3Host, s3AccessKey, s3SecretKey, false)
	if err != nil {
		panic(err)
	}

	var workerWg sync.WaitGroup
	remoteActionChan := make(chan backup.RemoteAction, 20)

	reportChan := make(chan backup.LogEntry)
	reportDone := make(chan struct{})
	reportOut := log.New(os.Stdout, "REPORT: ", log.Ldate|log.Ltime|log.LUTC)
	reportGenerator := backup.NewReporter(reportChan, reportDone, reportOut)
	go reportGenerator.Run()

	logger := backup.NewLogger(os.Stdout, reportChan, &workerWg)

	localFileProcessor := backup.NewLocalFileProcessor(targetDir)

	remoteFileProcessor, err := backup.NewRemoteFileProcessor(
		s3BucketName,
		s3Client.ListObjects,
		s3Client.RemoveObject,
		s3Client.FPutObject,
	)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		go worker.NewRemoteActionWorker(
			remoteFileProcessor.Put,
			remoteFileProcessor.Remove,
			&workerWg,
			remoteActionChan,
			logger,
		).Run()
	}

	processor := backup.NewProcessor(
		localFileProcessor.Gather,
		remoteFileProcessor.Gather,
		logger,
		&workerWg,
		remoteActionChan,
	)

	err = processor.Process()
	if err != nil {
		panic(err)
	}

	workerWg.Wait()
	reportDone <- struct{}{}
	reportGenerator.Print()
}

var targetDirViaFlag, targetDirViaEnv string
var s3HostViaFlag, s3HostViaEnv string
var s3AccessKeyViaFlag, s3AccessKeyViaEnv string
var s3SecretKeyViaFlag, s3SecretKeyViaEnv string
var s3BucketNameViaFlag, s3BucketNameViaEnv string

//FIXME There is probably a MUCH better way to do this.
func processVars() {
	flag.StringVar(&targetDirViaFlag, "targetDir", "", "Local directory to back up. Required.")
	flag.StringVar(&s3HostViaFlag, "s3Host", "", "S3 host. Required.")
	flag.StringVar(&s3AccessKeyViaFlag, "s3AccessKey", "", "S3 access key. Required.")
	flag.StringVar(&s3SecretKeyViaFlag, "s3SecretKey", "", "S3 secret key. Required.")
	flag.StringVar(&s3BucketNameViaFlag, "s3BucketName", "", "S3 Bucket Name. Optional.")
	flag.Parse()

	targetDirViaEnv = os.Getenv("PERSONAL_BACKUP_TARGET_DIR")
	s3HostViaEnv = os.Getenv("PERSONAL_BACKUP_S3_HOST")
	s3AccessKeyViaEnv = os.Getenv("PERSONAL_BACKUP_S3_ACCESS_KEY")
	s3SecretKeyViaFlag = os.Getenv("PERSONAL_BACKUP_S3_SECRET_KEY")
	s3BucketNameViaEnv = os.Getenv("PERSONAL_BACKUP_S3_BUCKET_NAME")

	if targetDirViaFlag != "" {
		targetDir = targetDirViaFlag
	} else if targetDirViaEnv != "" {
		targetDir = targetDirViaEnv
	} else {
		panic("target dir must be specified via either command line (-targetDir) or env var (PERSONAL_BACKUP_TARGET_DIR)")
	}

	if s3HostViaFlag != "" {
		s3Host = s3HostViaFlag
	} else if s3HostViaEnv != "" {
		s3Host = s3HostViaEnv
	} else {
		panic("s3 host must be specified via either command line (-s3Host) or env var (PERSONAL_BACKUP_S3_HOST)")
	}

	if s3AccessKeyViaFlag != "" {
		s3AccessKey = s3AccessKeyViaFlag
	} else if s3AccessKeyViaEnv != "" {
		s3AccessKey = s3AccessKeyViaEnv
	} else {
		panic("s3 access key must be specified via either command line (-s3AccessKey) or env var (PERSONAL_BACKUP_S3_ACCESS_KEY)")
	}

	if s3SecretKeyViaFlag != "" {
		s3SecretKey = s3SecretKeyViaFlag
	} else if s3SecretKeyViaEnv != "" {
		s3SecretKey = s3SecretKeyViaEnv
	} else {
		panic("s3 secret key must be specified via either command line (-s3SecretKey) or env var (PERSONAL_BACKUP_S3_SECRET_KEY)")
	}

	if s3BucketNameViaFlag != "" {
		s3BucketName = s3BucketNameViaFlag
	} else if s3BucketNameViaEnv != "" {
		s3BucketName = s3BucketNameViaEnv
	} else {
		panic("s3 bucket name must be specified via either command line (-s3BucketName) or env var (PERSONAL_BACKUP_S3_BUCKET_NAME)")
	}
}
