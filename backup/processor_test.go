package backup

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testLogger struct {
	logInfo  func(LogEntry)
	logError func(LogEntry)
}

func (l testLogger) Info(i LogEntry) {
	l.logInfo(i)
}

func (l testLogger) Error(i LogEntry) {
	l.logError(i)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

type ProcessorTestSuite struct {
	suite.Suite

	localGatherCalled, remoteGatherCalled     bool
	putToRemoteCalled, removeFromRemoteCalled bool

	localGather, remoteGather     func() (map[string]file, error)
	putToRemote, removeFromRemote func(string) error
	localData, remoteData         map[string]file

	logInfoCalled, logErrorCalled bool
	logger                        testLogger

	wg *sync.WaitGroup
}

func (s *ProcessorTestSuite) SetupTest() {
	s.localGatherCalled = false
	s.remoteGatherCalled = false
	s.putToRemoteCalled = false
	s.removeFromRemoteCalled = false

	s.localData = make(map[string]file)
	s.localData["local1"] = newFile("local1", 100)

	s.remoteData = make(map[string]file)
	s.remoteData["remote1"] = newFile("remote1", 100)

	s.localGather = func() (map[string]file, error) {
		s.localGatherCalled = true
		return s.localData, nil
	}

	s.remoteGather = func() (map[string]file, error) {
		s.remoteGatherCalled = true
		return s.remoteData, nil
	}

	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		return nil
	}

	s.removeFromRemote = func(f string) error {
		s.removeFromRemoteCalled = true
		return nil
	}

	s.logInfoCalled = false
	s.logErrorCalled = false

	s.logger = testLogger{
		logInfo: func(i LogEntry) {
			s.logInfoCalled = true
		},
		logError: func(i LogEntry) {
			s.logErrorCalled = true
		},
	}

	s.wg = &sync.WaitGroup{}
}

func (s ProcessorTestSuite) processor() processor {
	return NewProcessor(s.localGather, s.remoteGather, s.putToRemote, s.removeFromRemote, s.logger, s.wg)
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalGather() {
	s.processor().Process()
	s.True(s.localGatherCalled)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromLocalGather() {
	expectedErr := errors.New("asplode!")
	s.localGather = func() (map[string]file, error) {
		s.localGatherCalled = true
		return nil, expectedErr
	}

	s.logger.logError = func(i LogEntry) {
		s.logErrorCalled = true
		s.Equal(LogEntry{message: "error returned while gathering local files, err: asplode!"}, i)
	}

	err := s.processor().Process()

	s.wg.Wait()

	s.Require().True(s.localGatherCalled)
	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.True(s.logErrorCalled)
	s.False(s.logInfoCalled)
}

func (s *ProcessorTestSuite) Test_Process_CallsRatherGather() {
	s.processor().Process()
	s.True(s.remoteGatherCalled)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromRemoteGather() {
	expectedErr := errors.New("asplode!")
	s.remoteGather = func() (map[string]file, error) {
		s.remoteGatherCalled = true
		return nil, expectedErr
	}

	s.logger.logError = func(i LogEntry) {
		s.logErrorCalled = true
		s.Equal(LogEntry{message: "error returned while gathering remote files, err: asplode!"}, i)
	}

	err := s.processor().Process()

	s.wg.Wait()

	s.Require().Error(err)
	s.True(s.remoteGatherCalled)
	s.Equal(expectedErr, err)
	s.True(s.logErrorCalled)
	s.False(s.logInfoCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InBoth_Equal() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{"file": newFile("file", 100)}

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)

	s.False(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
	s.False(s.logErrorCalled)
	s.False(s.logInfoCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InBoth_NotEqual() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{"file": newFile("file", 101)}

	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		s.Equal("file", f)
		return nil
	}

	s.logger.logInfo = func(i LogEntry) {
		s.logInfoCalled = true
		s.Equal(
			LogEntry{
				message: fmt.Sprintf("mismatch, pushing to remote OLD - %s | NEW - %s", remote["file"], local["file"]),
				file:    "file",
			},
			i,
		)
	}

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)

	s.True(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
	s.True(s.logInfoCalled)
	s.False(s.logErrorCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InLocal_NotInRemote() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{}

	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		s.Equal("file", f)
		return nil
	}

	s.logger.logInfo = func(i LogEntry) {
		s.logInfoCalled = true
		s.Equal(
			LogEntry{
				message: fmt.Sprintf("not found, pushing to remote NEW - %s", local["file"]),
				file:    "file",
			},
			i,
		)
	}

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)

	s.True(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
	s.True(s.logInfoCalled)
	s.False(s.logErrorCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InLocal_NotInRemote_PushError() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{}

	expectedErr := errors.New("boom")
	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		return expectedErr
	}

	s.logger.logError = func(i LogEntry) {
		s.logErrorCalled = true
		s.Equal(
			LogEntry{
				message: fmt.Sprintf("unable to push to remote for file '%s', error: '%s'", local["file"], expectedErr.Error()),
				file:    "file",
			},
			i,
		)
	}

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)

	s.True(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
	s.False(s.logInfoCalled)
	s.True(s.logErrorCalled)
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InBoth() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{"file": newFile("file", 100)}

	s.wg.Add(1)
	s.processor().processRemoteVsLocal(local, remote)

	s.False(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InRemote_NotInLocal() {
	local := map[string]file{}
	remote := map[string]file{"file": newFile("file", 100)}

	s.removeFromRemote = func(f string) error {
		s.removeFromRemoteCalled = true
		s.Equal("file", f)
		return nil
	}

	s.logger.logInfo = func(i LogEntry) {
		s.logInfoCalled = true
		s.Equal(
			LogEntry{
				message: fmt.Sprintf("'%s' not found locally, removing from remote", remote["file"]),
				file:    "file",
			},
			i,
		)
	}

	s.wg.Add(1)
	s.processor().processRemoteVsLocal(local, remote)

	s.False(s.putToRemoteCalled)
	s.True(s.removeFromRemoteCalled)
	s.True(s.logInfoCalled)
	s.False(s.logErrorCalled)
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InRemote_NotInLocal_ErrorRemovingFromRemote() {
	local := map[string]file{}
	remote := map[string]file{"file": newFile("file", 100)}

	expectedErr := errors.New("AAAAAAAHHHHHHHHH")
	s.removeFromRemote = func(f string) error {
		s.removeFromRemoteCalled = true
		return expectedErr
	}

	s.logger.logError = func(i LogEntry) {
		s.logErrorCalled = true
		s.Equal(
			LogEntry{
				message: fmt.Sprintf("'%s' not found locally but unable to remove from remote, error: '%s'", remote["file"], expectedErr.Error()),
				file:    "file",
			},
			i,
		)
	}

	s.wg.Add(1)
	s.processor().processRemoteVsLocal(local, remote)

	s.False(s.putToRemoteCalled)
	s.True(s.removeFromRemoteCalled)
	s.False(s.logInfoCalled)
	s.True(s.logErrorCalled)
}

func (s *ProcessorTestSuite) Test_Process_MultipleDifferences() {
	local := map[string]file{
		"file1": newFile("file1", 100),
		"file2": newFile("file2", 200),
		"file3": newFile("file3", 300),
		"file4": newFile("file4", 400),
		"file5": newFile("file5", 500),
	}

	remote := map[string]file{
		"file1": newFile("file1", 100),
		"file2": newFile("file2", 201),
		"file3": newFile("file3", 300),
		"file4": newFile("file4", 400),
		"file6": newFile("file6", 600),
	}

	s.localGather = func() (map[string]file, error) {
		return local, nil
	}

	s.remoteGather = func() (map[string]file, error) {
		return remote, nil
	}

	putCalledCnt := 0
	s.putToRemote = func(f string) error {
		putCalledCnt++
		if f != "file2" && f != "file5" {
			s.Assert().Fail(fmt.Sprintf("Expected either 'file2' or 'file5' as put operation, received: '%s'", f))
		}
		return nil
	}

	removeCalledCnt := 0
	s.removeFromRemote = func(f string) error {
		removeCalledCnt++
		if removeCalledCnt == 1 {
			s.Equal("file6", f)
		}

		return nil
	}

	s.processor().Process()

	s.wg.Wait()

	s.Equal(2, putCalledCnt)
	s.Equal(1, removeCalledCnt)
}
