package backup

import (
	"sync"
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

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

type LoggerTestSuite struct {
	suite.Suite

	sliceLogger *sliceLogger

	report    chan LogEntry
	reportMsg LogEntry
	wg        *sync.WaitGroup

	logger BackupLogger
}

func (s *LoggerTestSuite) SetupTest() {
	s.sliceLogger = &sliceLogger{
		messages: make([]string, 0),
	}
	s.report = make(chan LogEntry)
	s.wg = &sync.WaitGroup{}
	s.logger = NewLogger(s.sliceLogger, s.report, s.wg)

	s.wg.Add(1)
	go func() {
		s.reportMsg = <-s.report
		s.wg.Done()
	}()
}

func (s *LoggerTestSuite) Test_Info_LogsToInfoLogger() {
	entry := LogEntry{message: "test", file: "testFile"}
	s.logger.Info(entry)

	s.wg.Wait()
	s.Contains(s.sliceLogger.messages[0], "INFO: ")
	s.Contains(s.sliceLogger.messages[0], "test")
	s.Contains(s.sliceLogger.messages[0], "testFile")
}

func (s *LoggerTestSuite) Test_Info_SendsEntryToReportChannel() {
	entry := LogEntry{message: "test", file: "testFile"}
	s.logger.Info(entry)

	s.wg.Wait()
	s.Equal(entry, s.reportMsg)
}

func (s *LoggerTestSuite) Test_Error_LogsToErrorLogger() {
	entry := LogEntry{message: "test", file: "testFile"}
	s.logger.Error(entry)

	s.wg.Wait()
	s.Contains(s.sliceLogger.messages[0], "ERROR: ")
	s.Contains(s.sliceLogger.messages[0], "test")
	s.Contains(s.sliceLogger.messages[0], "testFile")
}

func (s *LoggerTestSuite) Test_Error_SendsEntryToReportChannel() {
	entry := LogEntry{message: "test", file: "testFile"}
	s.logger.Error(entry)

	s.wg.Wait()
	s.Equal(entry, s.reportMsg)
}
