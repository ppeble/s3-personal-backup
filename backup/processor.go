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
	wg                *sync.WaitGroup
}

func NewProcessor(
	localGather, remoteGather func() (map[string]file, error),
	putToRemote, removeFromRemote func(string) error,
	log BackupLogger,
	wg *sync.WaitGroup,
) processor {
	return processor{
		gatherLocalFiles:  localGather,
		gatherRemoteFiles: remoteGather,
		putToRemote:       putToRemote,
		removeFromRemote:  removeFromRemote,
		logger:            log,
		wg:                wg,
	}
}

func (p processor) Process() (err error) {
	localResultChan := make(chan gatherResult)
	remoteResultChan := make(chan gatherResult)

	go func() {
		files, err := p.gatherLocalFiles()
		localResultChan <- gatherResult{result: files, err: err}
	}()

	go func() {
		files, err := p.gatherRemoteFiles()
		remoteResultChan <- gatherResult{result: files, err: err}
	}()

	local := <-localResultChan
	if local.err != nil {
		p.logger.Error(LogEntry{
			message: fmt.Sprintf("error returned while gathering local files, err: %s", local.err),
		})

		return local.err
	}

	//FIXME If the remote is somehow wrong then this hangs!
	remote := <-remoteResultChan
	if remote.err != nil {
		p.logger.Error(LogEntry{
			message: fmt.Sprintf("error returned while gathering remote files, err: %s", remote.err),
		})

		return remote.err
	}

	defer close(localResultChan)
	defer close(remoteResultChan)

	p.wg.Add(2)
	go p.processLocalVsRemote(local.result, remote.result)
	go p.processRemoteVsLocal(local.result, remote.result)

	return
}

func (p processor) processLocalVsRemote(local, remote map[string]file) {
	defer p.wg.Done()

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

func (p processor) processRemoteVsLocal(local, remote map[string]file) {
	defer p.wg.Done()

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
