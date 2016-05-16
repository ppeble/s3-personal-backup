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

	targetLocal, targetRemote string

	processor processor

	localGatherCalled, remoteGatherCalled bool

	localGather, remoteGather func() (map[string]file, error)
	localData, remoteData     map[string]file
}

func (s *ProcessorTestSuite) SetupTest() {
	s.targetLocal = "/tmp"
	s.targetRemote = "test"

	s.localGatherCalled = false
	s.remoteGatherCalled = false

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

	s.processor = NewProcessor(s.localGather, s.remoteGather)
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalGather() {
	s.processor.Process()
	s.True(s.localGatherCalled)
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromLocalGather() {
	expectedErr := errors.New("asplode!")
	localErrFunc := func() (map[string]file, error) {
		s.localGatherCalled = true
		return nil, expectedErr
	}

	err := NewProcessor(localErrFunc, s.remoteGather).Process()

	s.Require().True(s.localGatherCalled)
	s.Require().Error(err)
	s.Equal(expectedErr, err)
}

func (s *ProcessorTestSuite) Test_Process_CallsRatherGather() {
	s.processor.Process()
	s.True(s.remoteGatherCalled)
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromRemoteGather() {
	expectedErr := errors.New("asplode!")
	remoteErrFunc := func() (map[string]file, error) {
		return nil, expectedErr
	}

	err := NewProcessor(s.localGather, remoteErrFunc).Process()

	s.Require().Error(err)
	s.Equal(expectedErr, err)
}
