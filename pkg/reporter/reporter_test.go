package reporter

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ppeble/s3-personal-backup/pkg/backup"
)

func TestReporterTestSuite(t *testing.T) {
	suite.Run(t, new(ReporterTestSuite))
}

type ReporterTestSuite struct {
	suite.Suite

	sliceLogger *sliceLogger

	in     chan backup.LogEntry
	logger *log.Logger

	reporter reporter

	messageIterator int
}

func (s *ReporterTestSuite) SetupTest() {
	s.sliceLogger = &sliceLogger{
		messages: make([]string, 0),
	}

	s.in = make(chan backup.LogEntry)
	s.logger = log.New(s.sliceLogger, "REPORT: ", log.Ldate|log.Ltime|log.LUTC)
	s.reporter = NewReporter(s.in, s.logger)
	s.messageIterator = 0
}

func (s *ReporterTestSuite) Test_ReadsFromChannelAndLogs() {
	go s.reporter.Run()

	expectedEntry := backup.LogEntry{Message: "test", File: "file1"}
	s.in <- expectedEntry

	// Seems like it is possible for the 'Run' not getting the value in time
	time.Sleep(10 * time.Millisecond)

	s.Equal(s.reporter.entries[0], expectedEntry)
}

func (s *ReporterTestSuite) Test_Print_GeneratesReport() {
	go s.reporter.Run()

	s.in <- backup.LogEntry{Message: "test1", File: "file1", ActionType: backup.PUSH}
	s.in <- backup.LogEntry{Message: "test2", File: "file2", ActionType: backup.PUSH}
	s.in <- backup.LogEntry{Message: "test3", File: "file3", ActionType: backup.PUSH}
	s.in <- backup.LogEntry{Message: "test4", File: "file4", ActionType: backup.REMOVE}

	// Seems like it is possible for the 'Run' not getting the value in time
	time.Sleep(10 * time.Millisecond)

	s.reporter.Print()

	s.contains("Backup Report")
	s.contains("-------------------------------")
	s.contains("Total run time (in minutes): 0")
	s.contains("Total files processed: 4")
	s.contains("Time per file (in seconds):") // The time per file is highly variable
	s.contains("Files added to remote: 3")
	s.contains("Files removed from remote: 1")
	s.contains("")
	s.contains("File Details")
	s.contains("-------------------------------")
	s.contains("file: 'file1' - action: 'push' - message: 'test1'")
	s.contains("file: 'file2' - action: 'push' - message: 'test2'")
	s.contains("file: 'file3' - action: 'push' - message: 'test3'")
	s.contains("file: 'file4' - action: 'remove' - message: 'test4'")
	s.contains("")
}

func (s *ReporterTestSuite) contains(expected string) {
	s.Contains(s.sliceLogger.messages[s.messageIterator], expected)
	s.messageIterator++
}
