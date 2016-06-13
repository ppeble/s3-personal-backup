package worker

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ptrimble/dreamhost-personal-backup/backup"
)

func TestRemoteActionWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(RemoteActionWorkerTestSuite))
}

type RemoteActionWorkerTestSuite struct {
	suite.Suite

	putToRemoteCalled, removeFromRemoteCalled bool

	putToRemote, removeFromRemote func(string) error

	logInfoCalled, logErrorCalled bool
	logger                        testLogger

	input chan backup.RemoteAction
	wg    *sync.WaitGroup

	file backup.File
}

func (s *RemoteActionWorkerTestSuite) SetupTest() {
	s.file = backup.File{
		Name: "test1",
		Size: 100,
	}

	s.putToRemoteCalled = false
	s.removeFromRemoteCalled = false

	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		s.Equal(s.file.Name, f)
		return nil
	}

	s.removeFromRemote = func(f string) error {
		s.removeFromRemoteCalled = true
		s.Equal(s.file.Name, f)
		return nil
	}

	s.logInfoCalled = false
	s.logErrorCalled = false

	s.logger = testLogger{
		logInfo: func(i backup.LogEntry) {
			s.logInfoCalled = true
		},
		logError: func(i backup.LogEntry) {
			s.logErrorCalled = true
		},
	}

	s.input = make(chan backup.RemoteAction)
	s.wg = &sync.WaitGroup{}

	s.wg.Add(1)
}

func (s RemoteActionWorkerTestSuite) worker() RemoteActionWorker {
	return NewRemoteActionWorker(s.putToRemote, s.removeFromRemote, s.wg, s.input, s.logger)
}

func (s *RemoteActionWorkerTestSuite) Test_Run_HandlePush() {
	go s.worker().Run()

	s.input <- backup.RemoteAction{Type: backup.PUSH, File: s.file}

	s.True(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
	s.True(s.logInfoCalled)
}

func (s *RemoteActionWorkerTestSuite) Test_Run_Push_LogsErrorOnFailure() {
	s.putToRemote = func(f string) error {
		s.putToRemoteCalled = true
		return errors.New("asplode")
	}

	go s.worker().Run()

	s.input <- backup.RemoteAction{Type: backup.PUSH, File: s.file}

	s.True(s.putToRemoteCalled)
	s.False(s.removeFromRemoteCalled)
	s.False(s.logInfoCalled)
	s.True(s.logErrorCalled)
}

func (s *RemoteActionWorkerTestSuite) Test_Run_HandleRemove() {
	go s.worker().Run()

	s.input <- backup.RemoteAction{Type: backup.REMOVE, File: s.file}

	s.False(s.putToRemoteCalled)
	s.True(s.removeFromRemoteCalled)
	s.True(s.logInfoCalled)
}

//FIXME Flapping?
func (s *RemoteActionWorkerTestSuite) Test_Run_Remove_LogsErrorOnFailure() {
	s.removeFromRemote = func(f string) error {
		s.removeFromRemoteCalled = true
		return errors.New("asplode")
	}

	go s.worker().Run()

	s.input <- backup.RemoteAction{Type: backup.REMOVE, File: s.file}

	s.False(s.putToRemoteCalled)
	s.True(s.removeFromRemoteCalled)
	s.False(s.logInfoCalled)
	s.True(s.logErrorCalled)
}