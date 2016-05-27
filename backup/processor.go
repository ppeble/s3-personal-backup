package backup

import (
	"fmt"
	"sync"
)

type gatherResult struct {
	result map[string]file
	err    error
}

type processor struct {
	gatherLocalFiles  func() (map[string]file, error)
	gatherRemoteFiles func() (map[string]file, error)
	putToRemote       func(string) error
	removeFromRemote  func(string) error
	logger            BackupLogger
}

func NewProcessor(
	localGather, remoteGather func() (map[string]file, error),
	putToRemote, removeFromRemote func(string) error,
	log BackupLogger,
) processor {
	return processor{
		gatherLocalFiles:  localGather,
		gatherRemoteFiles: remoteGather,
		putToRemote:       putToRemote,
		removeFromRemote:  removeFromRemote,
		logger:            log,
	}
}

func (p processor) Process() (err error) {
	localResultChan := make(chan gatherResult)
	remoteResultChan := make(chan gatherResult)

	go func(out chan<- gatherResult) {
		files, err := p.gatherLocalFiles()
		out <- gatherResult{result: files, err: err}
	}(localResultChan)

	go func(out chan<- gatherResult) {
		files, err := p.gatherRemoteFiles()
		out <- gatherResult{result: files, err: err}
	}(remoteResultChan)

	local := <-localResultChan
	if local.err != nil {
		p.logger.Error(LogEntry{
			message: fmt.Sprintf("error returned while gathering local files, err: %s", local.err),
		})

		return local.err
	}

	remote := <-remoteResultChan
	if remote.err != nil {
		p.logger.Error(LogEntry{
			message: fmt.Sprintf("error returned while gathering remote files, err: %s", remote.err),
		})

		return remote.err
	}

	defer close(localResultChan)
	defer close(remoteResultChan)

	var wg sync.WaitGroup
	wg.Add(2)

	go p.processLocalVsRemote(local.result, remote.result, &wg)
	go p.processRemoteVsLocal(local.result, remote.result, &wg)

	wg.Wait()

	return
}

func (p processor) processLocalVsRemote(local, remote map[string]file, wg *sync.WaitGroup) {
	defer wg.Done()

	for lkey, lfile := range local {
		rfile, found := remote[lkey]

		var entry LogEntry
		push := false

		if !found {
			push = true
			entry = LogEntry{
				message: fmt.Sprintf("not found, pushing to remote NEW - %s", lfile),
				file:    lkey,
			}
		} else if !isEqual(lfile, rfile) {
			push = true
			entry = LogEntry{
				message: fmt.Sprintf("mismatch, pushing to remote OLD - %s | NEW - %s", rfile, lfile),
				file:    lkey,
			}
		}

		if push == true {
			err := p.putToRemote(lkey)
			if err != nil {
				p.logger.Error(LogEntry{
					message: fmt.Sprintf("unable to push to remote for file '%s', error: '%s'", lfile, err.Error()),
					file:    lkey,
				})
			} else {
				p.logger.Info(entry)
			}
		}
	}
}

func (p processor) processRemoteVsLocal(local, remote map[string]file, wg *sync.WaitGroup) {
	defer wg.Done()

	for rkey, rfile := range remote {
		_, found := local[rkey]
		if !found {
			err := p.removeFromRemote(rkey)
			if err != nil {
				entry := LogEntry{
					message: fmt.Sprintf("'%s' not found locally but unable to remove from remote, error: '%s'", rfile, err.Error()),
					file:    rkey,
				}
				p.logger.Error(entry)
			} else {
				entry := LogEntry{
					message: fmt.Sprintf("'%s' not found locally, removing from remote", rfile),
					file:    rkey,
				}
				p.logger.Info(entry)
			}
		}
	}
}
