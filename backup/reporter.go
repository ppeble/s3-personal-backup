package backup

import (
	"log"
	"sync"
)

func NewReporter(in <-chan LogEntry, done <-chan struct{}, l *log.Logger) reporter {
	return reporter{
		in:     in,
		done:   done,
		logger: l,
	}
}

type reporter struct {
	in     <-chan LogEntry
	done   <-chan struct{}
	logger *log.Logger
}

func (r *reporter) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case entry := <-r.in:
			r.logger.Println("file: '" + entry.file + "' - message: '" + entry.message + "'")
		case <-r.done:
			r.logger.Println("Received done signal, stopping reporting process")
			return
		}
	}
}
