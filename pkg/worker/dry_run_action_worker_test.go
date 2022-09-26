package worker

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ppeble/s3-personal-backup/pkg/backup"
)

func TestDryRunActionWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(DryRunActionWorkerTestSuite))
}

type DryRunActionWorkerTestSuite struct {
	suite.Suite

	file backup.File

	wg *sync.WaitGroup

	input  chan backup.RemoteAction
	report chan backup.LogEntry

	reportMsg backup.LogEntry
}

func (s *DryRunActionWorkerTestSuite) SetupTest() {
	s.file = backup.File{
		Name: "test1",
		Size: 100,
	}

	s.input = make(chan backup.RemoteAction)
	s.report = make(chan backup.LogEntry)

	s.wg = &sync.WaitGroup{}

	s.wg.Add(1)
	go func() {
		s.reportMsg = <-s.report
		s.wg.Done()
	}()
}

func (s DryRunActionWorkerTestSuite) worker() DryRunActionWorker {
	return NewDryRunActionWorker(s.wg, s.input, s.report)
}

func (s *DryRunActionWorkerTestSuite) Test_Run_SendsToReportChannel() {
	go s.worker().Run()

	s.wg.Add(1)
	s.input <- backup.RemoteAction{
		Type: backup.PUSH,
		File: s.file,
	}

	s.wg.Wait()

	s.Equal(
		backup.LogEntry{
			ActionType: backup.PUSH,
			File:       s.file.Name,
		},
		s.reportMsg,
	)
}
