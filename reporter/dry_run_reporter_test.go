package reporter

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ptrimble/dreamhost-personal-backup"
)

func TestDryRunReporterTestSuite(t *testing.T) {
	suite.Run(t, new(DryRunReporterTestSuite))
}

type DryRunReporterTestSuite struct {
	suite.Suite

	sliceLogger *sliceLogger

	in     chan backup.LogEntry
	logger *log.Logger

	reporter dryRunReporter

	messageIterator int
}

func (s *DryRunReporterTestSuite) SetupTest() {
	s.sliceLogger = &sliceLogger{
		messages: make([]string, 0),
	}

	s.in = make(chan backup.LogEntry)
	s.logger = log.New(s.sliceLogger, "REPORT: ", log.Ldate|log.Ltime|log.LUTC)
	s.reporter = NewDryRunReporter(s.in, s.logger)
	s.messageIterator = 0
}

func (s *DryRunReporterTestSuite) Test_ReadsFromChannelAndLogs() {
	go s.reporter.Run()

	expectedEntry := backup.LogEntry{Message: "test", File: "file1"}
	s.in <- expectedEntry

	// Seems like it is possible for the 'Run' not getting the value in time
	time.Sleep(10 * time.Millisecond)

	s.Equal(s.reporter.entries[0], expectedEntry)
}

func (s *DryRunReporterTestSuite) Test_Print_GeneratesReport() {
	go s.reporter.Run()

	s.in <- backup.LogEntry{Message: "test1", File: "file1", ActionType: backup.PUSH}
	s.in <- backup.LogEntry{Message: "test2", File: "file2", ActionType: backup.PUSH}
	s.in <- backup.LogEntry{Message: "test3", File: "file3", ActionType: backup.PUSH}
	s.in <- backup.LogEntry{Message: "test4", File: "file4", ActionType: backup.REMOVE}

	// Seems like it is possible for the 'Run' not getting the value in time
	time.Sleep(10 * time.Millisecond)

	s.reporter.Print()

	s.contains("Dry Run Report")
	s.contains("-------------------------------")
	s.contains("Total files processed: 4")
	s.contains("Files that would be added to remote: 3")
	s.contains("Files that would be removed from remote: 1")
	s.contains("")
	s.contains("File Details")
	s.contains("-------------------------------")
	s.contains("file: 'file1' - action: 'push' - message: 'test1'")
	s.contains("file: 'file2' - action: 'push' - message: 'test2'")
	s.contains("file: 'file3' - action: 'push' - message: 'test3'")
	s.contains("file: 'file4' - action: 'remove' - message: 'test4'")
	s.contains("")
}

func (s *DryRunReporterTestSuite) contains(expected string) {
	s.Contains(s.sliceLogger.messages[s.messageIterator], expected)
	s.messageIterator++
}
