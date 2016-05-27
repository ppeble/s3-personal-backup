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
	s.logger = log.New(s.sliceLogger, "INFO: ", log.Ldate|log.Ltime|log.LUTC)
	s.reporter = NewReporter(s.in, s.done, s.logger)
}

func (s *ReporterTestSuite) Test_Constructor() {
	s.Equal(
		reporter{
			in:     s.in,
			done:   s.done,
			logger: s.logger,
		},
		s.reporter,
	)
}

func (s *ReporterTestSuite) Test_ReadsFromChannelAndLogs() {
	var wg sync.WaitGroup
	wg.Add(1)
	go s.reporter.Run(&wg)
	s.in <- LogEntry{message: "test", file: "file1"}
	s.done <- struct{}{}

	wg.Wait()
	s.Contains(s.sliceLogger.messages[0], "INFO: ")
	s.Contains(s.sliceLogger.messages[0], "file: 'file1' - message: 'test'")
}

func (s *ReporterTestSuite) Test_ClosesReporterOnDone() {
	var wg sync.WaitGroup
	wg.Add(1)
	go s.reporter.Run(&wg)
	s.done <- struct{}{}

	wg.Wait()
	s.Contains(s.sliceLogger.messages[0], "INFO: ")
	s.Contains(s.sliceLogger.messages[0], "Received done signal, stopping reporting process")
}
