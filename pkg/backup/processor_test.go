package backup

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

type ProcessorTestSuite struct {
	suite.Suite

	localGatherCalled  int
	remoteGatherCalled bool

	localGatherers []FileGatherer
	remoteGatherer FileGatherer

	localGatherFunc  func() (FileData, error)
	remoteGatherFunc func() (FileData, error)

	localData, remoteData FileData

	logInfoCalled, logErrorCalled bool
	logger                        testLogger

	wg           *sync.WaitGroup
	remoteAction chan RemoteAction
}

func (s *ProcessorTestSuite) SetupTest() {
	s.localGatherCalled = 0
	s.remoteGatherCalled = false

	s.localData = make(FileData)
	s.localData["local1"] = newFile("local1", 100)

	s.remoteData = make(FileData)
	s.remoteData["remote1"] = newFile("remote1", 100)

	s.localGatherFunc = func() (FileData, error) {
		s.localGatherCalled++
		return s.localData, nil
	}

	s.localGatherers = make([]FileGatherer, 0)
	s.localGatherers = append(s.localGatherers, testGatherer{gather: s.localGatherFunc})

	s.remoteGatherFunc = func() (FileData, error) {
		s.remoteGatherCalled = true
		return s.remoteData, nil
	}
	s.remoteGatherer = testGatherer{gather: s.remoteGatherFunc}

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
	s.remoteAction = make(chan RemoteAction, 5)
}

func (s ProcessorTestSuite) processor() processor {
	return NewProcessor(s.localGatherers, s.remoteGatherer, s.logger, s.wg, s.remoteAction)
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalGather_OneLocalGather() {
	go func() {
		for {
			<-s.remoteAction
			s.wg.Done()
		}
	}()

	s.processor().Process()
	s.Equal(1, s.localGatherCalled)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalGather_MultipleLocalGathers() {
	go func() {
		for {
			<-s.remoteAction
			s.wg.Done()
		}
	}()

	s.localGatherers = []FileGatherer{
		testGatherer{
			gather: s.localGatherFunc,
		},
		testGatherer{
			gather: s.localGatherFunc,
		},
	}

	s.processor().Process()
	s.Equal(2, s.localGatherCalled)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromLocalGather() {
	expectedErr := errors.New("asplode!")
	s.localGatherFunc = func() (FileData, error) {
		s.localGatherCalled++
		return nil, expectedErr
	}

	s.localGatherers = make([]FileGatherer, 0)
	s.localGatherers = append(s.localGatherers, testGatherer{gather: s.localGatherFunc})

	s.logger.logError = func(i LogEntry) {
		s.logErrorCalled = true
		s.Equal(LogEntry{Message: "error returned while gathering local files, err: asplode!"}, i)
	}

	err := s.processor().Process()

	s.Require().Equal(1, s.localGatherCalled)
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
	s.remoteGatherFunc = func() (FileData, error) {
		s.remoteGatherCalled = true
		return nil, expectedErr
	}

	s.remoteGatherer = testGatherer{gather: s.remoteGatherFunc}

	s.logger.logError = func(i LogEntry) {
		s.logErrorCalled = true
		s.Equal(LogEntry{Message: "error returned while gathering remote files, err: asplode!"}, i)
	}

	err := s.processor().Process()

	s.Error(err)
	s.True(s.remoteGatherCalled)
	s.Equal(expectedErr, err)
	s.True(s.logErrorCalled)
	s.False(s.logInfoCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InBoth_Equal() {
	local := FileData{"file": newFile("file", 100)}
	remote := FileData{"file": newFile("file", 100)}

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)

	s.False(s.logErrorCalled)
	s.False(s.logInfoCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InBoth_NotEqual() {
	local := FileData{"file": newFile("file", 100)}
	remote := FileData{"file": newFile("file", 101)}

	go func() {
		action := <-s.remoteAction
		s.Equal(ActionType(PUSH), action.Type)
		s.Equal("file", action.File.Name)
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InLocal_NotInRemote() {
	local := FileData{"file": newFile("file", 100)}
	remote := FileData{}

	go func() {
		action := <-s.remoteAction
		s.Equal(ActionType(PUSH), action.Type)
		s.Equal("file", action.File.Name)
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.processor().processLocalVsRemote(local, remote)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InBoth() {
	local := FileData{"file": newFile("file", 100)}
	remote := FileData{"file": newFile("file", 100)}

	s.wg.Add(1)
	s.processor().processRemoteVsLocal(local, remote)
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InRemote_NotInLocal() {
	local := FileData{}
	remote := FileData{"file": newFile("file", 100)}

	go func() {
		action := <-s.remoteAction
		s.Equal(ActionType(REMOVE), action.Type)
		s.Equal("file", action.File.Name)
		s.wg.Done()
	}()

	s.wg.Add(1)
	s.processor().processRemoteVsLocal(local, remote)
	s.wg.Wait()
}

func (s *ProcessorTestSuite) Test_Process_MultipleDifferences_SingleLocal() {
	local := FileData{
		"file1": newFile("file1", 100),
		"file2": newFile("file2", 200),
		"file3": newFile("file3", 300),
		"file4": newFile("file4", 400),
		"file5": newFile("file5", 500),
	}

	remote := FileData{
		"file1": newFile("file1", 100),
		"file2": newFile("file2", 201),
		"file3": newFile("file3", 300),
		"file4": newFile("file4", 400),
		"file6": newFile("file6", 600),
	}

	s.localGatherFunc = func() (FileData, error) {
		return local, nil
	}

	s.localGatherers = make([]FileGatherer, 0)
	s.localGatherers = append(s.localGatherers, testGatherer{gather: s.localGatherFunc})

	s.remoteGatherFunc = func() (FileData, error) {
		return remote, nil
	}

	s.remoteGatherer = testGatherer{gather: s.remoteGatherFunc}

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

	s.Equal(2, putCalledCnt, "putCalledCnt does not match")
	s.Equal(1, removeCalledCnt, "removeCalledCnt does not match")
}

func (s *ProcessorTestSuite) Test_Process_MultipleDifferences_MultipleLocal() {
	local1 := FileData{
		"local1/file1": newFile("local1/file1", 100),
		"local1/file2": newFile("local1/file2", 200),
		"local1/file3": newFile("local1/file3", 300),
		"local1/file4": newFile("local1/file4", 400),
		"local1/file5": newFile("local1/file5", 500),
	}

	local2 := FileData{
		"local2/file1": newFile("local2/file1", 100),
		"local2/file2": newFile("local2/file2", 200),
	}

	remote := FileData{
		"local1/file1": newFile("local1/file1", 100),
		"local1/file2": newFile("local1/file2", 201),
		"local1/file3": newFile("local1/file3", 300),
		"local1/file4": newFile("local1/file4", 400),
		"local1/file6": newFile("local1/file6", 600),
		"local2/file1": newFile("local2/file1", 100),
		"local2/file2": newFile("local2/file2", 201),
		"local2/file3": newFile("local2/file3", 300),
	}

	s.localGatherers = []FileGatherer{
		testGatherer{
			gather: func() (FileData, error) {
				return local1, nil
			},
		},
		testGatherer{
			gather: func() (FileData, error) {
				return local2, nil
			},
		},
	}

	s.remoteGatherFunc = func() (FileData, error) {
		return remote, nil
	}

	s.remoteGatherer = testGatherer{gather: s.remoteGatherFunc}

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

	s.Equal(3, putCalledCnt, "putCalledCnt does not match")
	s.Equal(2, removeCalledCnt, "removeCalledCnt does not match")
}

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

type testGatherer struct {
	gather func() (FileData, error)
}

func (g testGatherer) Gather() (FileData, error) {
	return g.gather()
}
