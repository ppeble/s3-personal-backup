package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/minio/minio-go"

	"github.com/ptrimble/dreamhost-personal-backup/backup"
	"github.com/ptrimble/dreamhost-personal-backup/backup/logger"
	"github.com/ptrimble/dreamhost-personal-backup/backup/worker"
)

func main() {
	config := processVars()

	s3Client, err := minio.NewV2(config.S3Host, config.S3AccessKey, config.S3SecretKey, false)
	if err != nil {
		panic(err)
	}

	var workerWg sync.WaitGroup
	remoteActionChan := make(chan backup.RemoteAction, 20)

	reportChan := make(chan logger.LogEntry)
	reportDone := make(chan struct{})
	reportOut := log.New(os.Stdout, "REPORT: ", log.Ldate|log.Ltime|log.LUTC)
	reportGenerator := backup.NewReporter(reportChan, reportDone, reportOut)
	go reportGenerator.Run()

	logger := logger.NewLogger(os.Stdout, reportChan, &workerWg)

	localFileProcessor := backup.NewLocalFileProcessor(config.TargetDir)

	remoteFileProcessor, err := backup.NewRemoteFileProcessor(
		config.S3BucketName,
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

func processVars() backup.CompiledConfig {
	flags := backup.Flags{}

	flag.StringVar(&flags.TargetDir, "targetDir", "", "Local directory to back up.")
	flag.StringVar(&flags.S3Host, "s3Host", "", "S3 host.")
	flag.StringVar(&flags.S3AccessKey, "s3AccessKey", "", "S3 access key.")
	flag.StringVar(&flags.S3SecretKey, "s3SecretKey", "", "S3 secret key.")
	flag.StringVar(&flags.S3BucketName, "s3BucketName", "", "S3 Bucket Name.")
	flag.Parse()

	compiledConfig, err := backup.CompileConfig(flags)
	if err != nil {
		panic(err)
	}

	return compiledConfig
}
