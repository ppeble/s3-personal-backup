package step_definitions

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	. "github.com/gucumber/gucumber"
	"github.com/minio/minio-go"
	"github.com/spf13/viper"

	"github.com/ptrimble/dreamhost-personal-backup"
	"github.com/ptrimble/dreamhost-personal-backup/logger"
	"github.com/ptrimble/dreamhost-personal-backup/reporter"
	"github.com/ptrimble/dreamhost-personal-backup/worker"
)

var SetupWebSteps = setupWebSteps()
var s3Client *minio.Client
var remoteActionChan chan backup.RemoteAction
var reportGenerator backup.Reporter
var outLogger backupLogger
var reportChan chan backup.LogEntry
var workerWg sync.WaitGroup
var out Output

type Output struct {
	out []string
}

func (w *Output) Write(b []byte) (int, error) {
	w.out = append(w.out, string(b[:]))
	return len(b), nil
}

type backupLogger interface {
	Info(backup.LogEntry)
	Error(backup.LogEntry)
}

func setupWebSteps() bool {
	Before("", func() {
		processVars()

		workerWg = sync.WaitGroup{}

		createS3Client()
		remoteActionChan = make(chan backup.RemoteAction, 20)
	})

	Given(`^this is not a dry run$`, func() {
		reportChan = make(chan backup.LogEntry)
		reportOut := log.New(&out, "REPORT: ", log.Ldate|log.Ltime|log.LUTC)
		r := reporter.NewReporter(reportChan, reportOut)
		reportGenerator = &r
		go reportGenerator.Run()

		outLogger = logger.NewLogger(&out, reportChan, &workerWg)
		remoteFileProcessor := remoteFileProcessor()

		for i := 0; i < viper.GetInt("remoteWorkerCount"); i++ {
			go worker.NewRemoteActionWorker(
				remoteFileProcessor.Put,
				remoteFileProcessor.Remove,
				&workerWg,
				remoteActionChan,
				outLogger,
			).Run()
		}
	})

	When(`^I run the backup for a single directory$`, func() {
		targetDir, err := ioutil.TempDir("/tmp", "")
		if err != nil {
			panic(err) //TODO Should this be a test failure?
		}

		createTempFile(targetDir, "tmpFile1")
		createTempFile(targetDir, "tempFile2")
		createTempFile(targetDir, "tempFile3")

		lfps := make([]backup.FileGatherer, 1)
		p := backup.NewLocalFileProcessor(targetDir)
		lfps[0] = &p

		rp := remoteFileProcessor()

		processor := backup.NewProcessor(
			lfps,
			&rp,
			outLogger,
			&workerWg,
			remoteActionChan,
		)

		err = processor.Process()
		if err != nil {
			panic(err) // TODO Should this be a test failure?
		}

		workerWg.Wait()
		reportGenerator.Print()
	})

	Then(`^I should see the expected files on the s(\d+) host$`, func(i1 int) {
		fileData := make(backup.FileData)

		// Create a done channel to control 'ListObjects' go routine.
		doneCh := make(chan struct{})

		// Indicate to our routine to exit cleanly upon return.
		defer close(doneCh)

		objectCh := s3Client.ListObjects(viper.GetString("s3BucketName"), "", true, doneCh)
		for object := range objectCh {
			if object.Err != nil {
				panic(object.Err)
			}

			fileData[backup.Filename(object.Key)] = newFile(object.Key, object.Size)
		}
		fmt.Printf("fileData: %#v\n", fileData)
	})

	And(`^I should see the following output:$`, func(data string) {

	})

	return true
}

func createS3Client() {
	var err error
	s3Client, err = minio.New(
		viper.GetString("s3Host"),
		viper.GetString("s3AccessKey"),
		viper.GetString("s3SecretKey"),
		false, // Mark as http since we will be testing locally
	)
	if err != nil {
		panic(err) // TODO Should this be a test failure?
	}
}

func remoteFileProcessor() backup.RemoteFileProcessor {
	p, err := backup.NewRemoteFileProcessor(
		viper.GetString("s3BucketName"),
		s3Client.ListObjects,
		s3Client.RemoveObject,
		s3Client.FPutObject,
	)
	if err != nil {
		panic(err)
	}

	return p
}

func processVars() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("PERSONAL_BACKUP")
	viper.BindEnv("targetDirs")
	viper.BindEnv("s3Host")
	viper.BindEnv("s3AccessKey")
	viper.BindEnv("s3SecretKey")
	viper.BindEnv("s3BucketName")
	viper.BindEnv("remoteWorkerCount")

	viper.SetDefault("remoteWorkerCount", 5)
}

func createTempFile(directory, prefix string) *os.File {
	tmpFile, err := ioutil.TempFile(directory, prefix)
	if err != nil {
		panic(err) //TODO Should this be a test failure?
	}

	return tmpFile
}

func newFile(name string, size int64) backup.File {
	return backup.File{
		Name: name,
		Size: size,
	}
}
