package backup

import (
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestReporterTestSuite(t *testing.T) {
	suite.Run(t, new(ReporterTestSuite))
}

type ReporterTestSuite struct {
	suite.Suite

	sliceLogger *sliceLogger

	in     chan LogEntry
	done   chan struct{}
	logger *log.Logger

	reporter reporter
}

func (s *ReporterTestSuite) SetupTest() {
	s.sliceLogger = &sliceLogger{
		messages: make([]string, 0),
	}

	s.in = make(chan LogEntry)
	s.done = make(chan struct{})
	s.logger = log.New(s.sliceLogger, "REPORT: ", log.Ldate|log.Ltime|log.LUTC)
	s.reporter = NewReporter(s.in, s.done, s.logger)
}

func (s *ReporterTestSuite) Test_ReadsFromChannelAndLogs() {
	var wg sync.WaitGroup
	wg.Add(1)
	go s.reporter.Run(&wg)

	expectedEntry := LogEntry{message: "test", file: "file1"}
	s.in <- expectedEntry
	s.done <- struct{}{}

	wg.Wait()
	s.Equal(s.reporter.entries[0], expectedEntry)
}

func (s *ReporterTestSuite) Test_ClosesReporterOnDone() {
	var wg sync.WaitGroup
	wg.Add(1)
	go s.reporter.Run(&wg)
	s.done <- struct{}{}

	wg.Wait()
	s.Contains(s.sliceLogger.messages[0], "Received done signal, stopping reporting process")
}

func (s *ReporterTestSuite) Test_Print_GeneratesReport() {
	var wg sync.WaitGroup
	wg.Add(1)
	go s.reporter.Run(&wg)

	s.in <- LogEntry{message: "test1", file: "file1"}
	s.in <- LogEntry{message: "test2", file: "file2"}
	s.in <- LogEntry{message: "test3", file: "file3"}
	s.done <- struct{}{}

	wg.Wait()

	s.reporter.Print()

	s.Contains(s.sliceLogger.messages[1], "Report")
	s.Contains(s.sliceLogger.messages[2], "-------------------------------")
	s.Contains(s.sliceLogger.messages[3], "file: 'file1' - message: 'test1'")
	s.Contains(s.sliceLogger.messages[4], "file: 'file2' - message: 'test2'")
	s.Contains(s.sliceLogger.messages[5], "file: 'file3' - message: 'test3'")
	s.Contains(s.sliceLogger.messages[6], "")
}
