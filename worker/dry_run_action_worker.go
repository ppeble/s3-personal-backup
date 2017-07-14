package worker

import (
	"sync"

	"github.com/ppeble/dreamhost-personal-backup"
)

type DryRunActionWorker struct {
	wg     *sync.WaitGroup
	in     <-chan backup.RemoteAction
	report chan<- backup.LogEntry
}

func NewDryRunActionWorker(wg *sync.WaitGroup, in <-chan backup.RemoteAction, report chan<- backup.LogEntry) DryRunActionWorker {
	return DryRunActionWorker{
		wg:     wg,
		in:     in,
		report: report,
	}
}

func (w DryRunActionWorker) Run() {
	for {
		action := <-w.in

		w.report <- backup.LogEntry{
			File:       action.File.Name,
			ActionType: action.Type,
		}

		w.wg.Done()
	}
}
