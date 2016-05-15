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

	localGather, remoteGather func(string) (map[string]file, error)
}

func (s *ProcessorTestSuite) SetupTest() {
	s.targetLocal = "/tmp"
	s.targetRemote = "test"

	s.localGatherCalled = false
	s.remoteGatherCalled = false

	s.localGather = func(input string) (map[string]file, error) {
		s.Equal(s.targetLocal, input)
		s.localGatherCalled = true
		return nil, nil
	}

	s.remoteGather = func(input string) (map[string]file, error) {
		s.Equal(s.targetRemote, input)
		s.remoteGatherCalled = true
		return nil, nil
	}

	s.processor = NewProcessor(s.localGather, s.remoteGather)
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalGather() {
	s.processor.Process(s.targetLocal, s.targetRemote)
	s.True(s.localGatherCalled)
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromLocalGather() {
	expectedErr := errors.New("asplode!")
	localErrFunc := func(input string) (map[string]file, error) {
		s.localGatherCalled = true
		return nil, expectedErr
	}

	err := NewProcessor(localErrFunc, s.remoteGather).Process(s.targetLocal, s.targetRemote)

	s.Require().True(s.localGatherCalled)
	s.Require().Error(err)
	s.Equal(expectedErr, err)
}

//TODO Must call remote and test remote error
//TODO Then test that the right data comes back from both and let's print it.
//TODO then let's test the command line, compare remote contents and local contents
