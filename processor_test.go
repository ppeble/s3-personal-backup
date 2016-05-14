package backup

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

type ProcessorTestSuite struct {
	suite.Suite

	targetDir string

	processor processor

	localProcessorCalled bool
	localProcessor       func(string) (map[string][]os.FileInfo, error)
}

func (s *ProcessorTestSuite) SetupTest() {
	s.targetDir = "/tmp"

	s.localProcessor = func(input string) (map[string][]os.FileInfo, error) {
		s.Equal(s.targetDir, input)
		s.localProcessorCalled = true
		return nil, nil
	}

	s.processor = NewProcessor(s.localProcessor)
}

func (s *ProcessorTestSuite) Test_Process_CallsLocalProcessor() {
	s.processor.Process(s.targetDir)
	s.True(s.localProcessorCalled)
}

func (s *ProcessorTestSuite) Test_Process_ReturnsErrorFromLocalProcessor() {
	expectedErr := errors.New("asplode!")
	localErrFunc := func(input string) (map[string][]os.FileInfo, error) {
		s.localProcessorCalled = true
		return nil, expectedErr
	}

	err := NewProcessor(localErrFunc).Process(s.targetDir)

	s.Require().True(s.localProcessorCalled)
	s.Require().Error(err)
	s.Equal(expectedErr, err)
}
