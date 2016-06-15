package backup

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ptrimble/dreamhost-personal-backup/backup/logger"
)

type testLogger struct {
	logInfo  func(logger.LogEntry)
	logError func(logger.LogEntry)
}

func (l testLogger) Info(i logger.LogEntry) {
	l.logInfo(i)
}

func (l testLogger) Error(i logger.LogEntry) {
	l.logError(i)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

type ProcessorTestSuite struct {
	suite.Suite

	localGatherCalled, remoteGatherCalled bool

	localGather, remoteGather func() (map[string]File, error)
	localData, remoteData     map[string]File

	logInfoCalled, logErrorCalled bool
	logger                        testLogger

	wg           *sync.WaitGroup
	remoteAction chan RemoteAction
}

func (s *ProcessorTestSuite) SetupTest() {
	s.localGatherCalled = false
	s.remoteGatherCalled = false

	s.localData = make(map[string]File)
	s.localData["local1"] = newFile("local1", 100)

	s.remoteData = make(map[string]File)
	s.remoteData["remote1"] = newFile("remote1", 100)

	s.localGather = func() (map[string]File, error) {
		s.localGatherCalled = true
		return s.localData, nil
	}

	s.remoteGather = func() (map[string]File, error) {
		s.remoteGatherCalled = true
		return s.remoteData, nil
	}

	s.logInfoCalled = false
	s.logErrorCalled = false

	s.logger = testLogger{
		logInfo: func(i logger.LogEntry) {
			s.logInfoCalled = true
		},
		logError: func(i logger.LogEntry) {
			s.logErrorCalled = true
		},
	}

	s.wg = &sync.WaitGroup{}
	s.remoteAction = make(chan RemoteAction, 5)
}

func (s ProcessorTestSuite) processor() processor {
	return NewProcessor(s.localGather, s.remoteGather, s.logger, s.wg, s.remoteAction)
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalGather() {
	go func() {
		for {
			<-s.remoteAction
			s.wg.Done()
		}
	}()

	s.processor().Process()
	s.True(s.localGatherCalled)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromLocalGather() {
	expectedErr := errors.New("asplode!")
	s.localGather = func() (map[string]File, error) {
		s.localGatherCalled = true
		return nil, expectedErr
	}

	s.logger.logError = func(i logger.LogEntry) {
		s.logErrorCalled = true
		s.Equal(logger.LogEntry{Message: "error returned while gathering local files, err: asplode!"}, i)
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
	go func() {
		for {
			<-s.remoteAction
			s.wg.Done()
		}
	}()

	s.processor().Process()
	s.True(s.remoteGatherCalled)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromRemoteGather() {
	expectedErr := errors.New("asplode!")
	s.remoteGather = func() (map[string]File, error) {
		s.remoteGatherCalled = true
		return nil, expectedErr
	}

	s.logger.logError = func(i logger.LogEntry) {
		s.logErrorCalled = true
		s.Equal(logger.LogEntry{Message: "error returned while gathering remote files, err: asplode!"}, i)
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
	local := map[string]File{"file": newFile("file", 100)}
	remote := map[string]File{"file": newFile("file", 100)}

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)

	s.False(s.logErrorCalled)
	s.False(s.logInfoCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InBoth_NotEqual() {
	local := map[string]File{"file": newFile("file", 100)}
	remote := map[string]File{"file": newFile("file", 101)}

	go func() {
		action := <-s.remoteAction
		s.Equal(PUSH, action.Type)
		s.Equal("file", action.File.Name)
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InLocal_NotInRemote() {
	local := map[string]File{"file": newFile("file", 100)}
	remote := map[string]File{}

	go func() {
		action := <-s.remoteAction
		s.Equal(PUSH, action.Type)
		s.Equal("file", action.File.Name)
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InBoth() {
	local := map[string]File{"file": newFile("file", 100)}
	remote := map[string]File{"file": newFile("file", 100)}

	s.wg.Add(1)
	s.processor().processRemoteVsLocal(local, remote)
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InRemote_NotInLocal() {
	local := map[string]File{}
	remote := map[string]File{"file": newFile("file", 100)}

	go func() {
		action := <-s.remoteAction
		s.Equal(REMOVE, action.Type)
		s.Equal("file", action.File.Name)
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.processor().processRemoteVsLocal(local, remote)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_MultipleDifferences() {
	local := map[string]File{
		"file1": newFile("file1", 100),
		"file2": newFile("file2", 200),
		"file3": newFile("file3", 300),
		"file4": newFile("file4", 400),
		"file5": newFile("file5", 500),
	}

	remote := map[string]File{
		"file1": newFile("file1", 100),
		"file2": newFile("file2", 201),
		"file3": newFile("file3", 300),
		"file4": newFile("file4", 400),
		"file6": newFile("file6", 600),
	}

	s.localGather = func() (map[string]File, error) {
		return local, nil
	}

	s.remoteGather = func() (map[string]File, error) {
		return remote, nil
	}

	putCalledCnt := 0
	removeCalledCnt := 0
	go func() {
		for {
			action := <-s.remoteAction
			if action.Type == PUSH {
				putCalledCnt++
			} else if action.Type == REMOVE {
				removeCalledCnt++
			} else {
				s.T().Errorf("Unknown action type")
			}
			s.wg.Done()
		}
	}()

	s.processor().Process()

	s.wg.Wait()

	s.Equal(2, putCalledCnt)
	s.Equal(1, removeCalledCnt)
}
