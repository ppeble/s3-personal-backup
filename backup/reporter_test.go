package backup

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"
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

	in     chan logEntry
	done   chan struct{}
	logger *log.Logger

	reporter reporter
}

func (s *ReporterTestSuite) SetupTest() {
	s.sliceLogger = &sliceLogger{
		messages: make([]string, 0),
	}

	s.in = make(chan logEntry)
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
	go s.reporter.Run()
	s.in <- logEntry{message: "test", file: "file1"}

	s.Contains(s.sliceLogger.messages[0], "INFO: ")
	s.Contains(s.sliceLogger.messages[0], "file: 'file1' - message: 'test'")
}

func (s *ReporterTestSuite) Test_ClosesReporterOnDone() {
	go s.reporter.Run()
	s.done <- struct{}{}

	s.Contains(s.sliceLogger.messages[0], "INFO: ")
	s.Contains(s.sliceLogger.messages[0], "Received done signal, stopping reporting process")
}
