package backup

import (
	"log"
)

func NewReporter(
	in <-chan LogEntry,
	done <-chan struct{},
	l *log.Logger,
) reporter {
	return reporter{
		in:      in,
		done:    done,
		logger:  l,
		entries: make([]LogEntry, 0),
	}
}

type reporter struct {
	in     <-chan LogEntry
	done   <-chan struct{}
	logger *log.Logger

	entries []LogEntry
}

func (r *reporter) Run() {
	for {
		select {
		case entry := <-r.in:
			r.entries = append(r.entries, entry)
		case <-r.done:
			r.logger.Println("Received done signal, waiting for all processes to finish")
			return
		}
	}
}

//TODO Add some kind of timestamp in here, this is what we will probably want to be
// printed to a separate file, it'll be nice to have some indication
func (r *reporter) Print() {
	r.logger.Println("Report")
	r.logger.Println("-------------------------------")

	for _, entry := range r.entries {
		r.logger.Println(entry.String())
	}

	r.logger.Println("")
}
