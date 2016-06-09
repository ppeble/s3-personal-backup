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
	actionChan := make(chan file, 10)

	for i := 0; i < 5; i++ {
		go p.remotePushWorker(actionChan)
	}

	for lkey, lfile := range local {
		rfile, found := remote[lkey]
		if !found || !isEqual(lfile, rfile) {
			p.wg.Add(1)
			actionChan <- lfile
		}
	}
}

func (p processor) remotePushWorker(in <-chan file) {
	for {
		file := <-in
		err := p.putToRemote(file.name)
		if err != nil {
			p.logger.Error(LogEntry{
				message: fmt.Sprintf("unable to push to remote for file '%s', error: '%s'", file, err.Error()),
				file:    file.name,
			})
		} else {
			p.logger.Info(LogEntry{
				message: fmt.Sprintf("%s pushed to remote", file),
				file:    file.name,
			})
		}

		p.wg.Done()
	}
}

func (p processor) processRemoteVsLocal(local, remote map[string]file) {
	defer p.wg.Done()
	actionChan := make(chan file, 10)

	for i := 0; i < 5; i++ {
		go p.remoteRemoveWorker(actionChan)
	}

	for rkey, rfile := range remote {
		_, found := local[rkey]
		if !found {
			p.wg.Add(1)
			actionChan <- rfile
		}
	}
}

func (p processor) remoteRemoveWorker(in <-chan file) {
	for {
		file := <-in
		err := p.removeFromRemote(file.name)
		if err != nil {
			entry := LogEntry{
				message: fmt.Sprintf("%s not found locally but unable to remove from remote, error: '%s'", file, err.Error()),
				file:    file.name,
			}
			p.logger.Error(entry)
		} else {
			entry := LogEntry{
				message: fmt.Sprintf("%s not found locally, removing from remote", file),
				file:    file.name,
			}
			p.logger.Info(entry)
		}

		p.wg.Done()
	}

}
