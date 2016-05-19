package backup

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

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
}

func (s ProcessorTestSuite) processor() processor {
	return NewProcessor(s.localGather, s.remoteGather, s.putToRemote, s.removeFromRemote)
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalGather() {
	s.processor().Process()
	s.True(s.localGatherCalled)
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromLocalGather() {
	expectedErr := errors.New("asplode!")
	s.localGather = func() (map[string]file, error) {
		s.localGatherCalled = true
		return nil, expectedErr
	}

	err := s.processor().Process()

	s.Require().True(s.localGatherCalled)
	s.Require().Error(err)
	s.Equal(expectedErr, err)
}

func (s *ProcessorTestSuite) Test_Process_CallsRatherGather() {
	s.processor().Process()
	s.True(s.remoteGatherCalled)
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromRemoteGather() {
	expectedErr := errors.New("asplode!")
	s.remoteGather = func() (map[string]file, error) {
		return nil, expectedErr
	}

	err := s.processor().Process()

	s.Require().Error(err)
	s.Equal(expectedErr, err)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InBoth_Equal() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{"file": newFile("file", 100)}

	s.processor().processLocalVsRemote(local, remote)

	s.False(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InBoth_NotEqual() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{"file": newFile("file", 101)}

	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		s.Equal("file", f)
		return nil
	}

	s.processor().processLocalVsRemote(local, remote)

	s.True(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
}

func (s *ProcessorTestSuite) Test_processLocalVsRemote_InLocal_NotInRemote() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{}

	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		s.Equal("file", f)
		return nil
	}

	s.processor().processLocalVsRemote(local, remote)

	s.True(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
}

func (s *ProcessorTestSuite) Test_processRemoteVsLocal_InBoth() {
	local := map[string]file{"file": newFile("file", 100)}
	remote := map[string]file{"file": newFile("file", 100)}

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

	s.processor().processRemoteVsLocal(local, remote)

	s.False(s.putToRemoteCalled)
	s.True(s.removeFromRemoteCalled)
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
		if putCalledCnt == 1 {
			s.Equal("file2", f)
		} else if putCalledCnt == 2 {
			s.Equal("file5", f)
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

	s.Equal(2, putCalledCnt)
	s.Equal(1, removeCalledCnt)
}
