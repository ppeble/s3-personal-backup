package backup

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ptrimble/dreamhost-personal-backup/backup/logger"
)

type sliceLogger struct {
	messages []string
}

func (l *sliceLogger) Write(b []byte) (n int, err error) {
	msg := string(b[:])
	l.messages = append(l.messages, msg)
	return len(msg), nil
}

func TestReporterTestSuite(t *testing.T) {
	suite.Run(t, new(ReporterTestSuite))
}

type ReporterTestSuite struct {
	suite.Suite

	sliceLogger *sliceLogger

	in     chan logger.LogEntry
	done   chan struct{}
	logger *log.Logger

	reporter reporter
}

func (s *ReporterTestSuite) SetupTest() {
	s.sliceLogger = &sliceLogger{
		messages: make([]string, 0),
	}

	s.in = make(chan logger.LogEntry)
	s.done = make(chan struct{})
	s.logger = log.New(s.sliceLogger, "REPORT: ", log.Ldate|log.Ltime|log.LUTC)
	s.reporter = NewReporter(s.in, s.done, s.logger)
}

func (s *ReporterTestSuite) Test_ReadsFromChannelAndLogs() {
	go s.reporter.Run()

	expectedEntry := logger.LogEntry{Message: "test", File: "file1"}
	s.in <- expectedEntry

	s.done <- struct{}{}
	s.Equal(s.reporter.entries[0], expectedEntry)
}

func (s *ReporterTestSuite) Test_ClosesReporterOnDone() {
	go s.reporter.Run()
	s.done <- struct{}{}
	time.Sleep(10 * time.Millisecond) // Need to give Run() time to complete

	s.Contains(s.sliceLogger.messages[0], "Received done signal, waiting for all processes to finish")
}

func (s *ReporterTestSuite) Test_Print_GeneratesReport() {
	go s.reporter.Run()

	s.in <- logger.LogEntry{Message: "test1", File: "file1"}
	s.in <- logger.LogEntry{Message: "test2", File: "file2"}
	s.in <- logger.LogEntry{Message: "test3", File: "file3"}

	s.done <- struct{}{}
	time.Sleep(10 * time.Millisecond) // Need to give Run() time to complete

	s.reporter.Print()

	s.Contains(s.sliceLogger.messages[1], "Report")
	s.Contains(s.sliceLogger.messages[2], "-------------------------------")
	s.Contains(s.sliceLogger.messages[3], "file: 'file1' - message: 'test1'")
	s.Contains(s.sliceLogger.messages[4], "file: 'file2' - message: 'test2'")
	s.Contains(s.sliceLogger.messages[5], "file: 'file3' - message: 'test3'")
	s.Contains(s.sliceLogger.messages[6], "")
}
