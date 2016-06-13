package backup

import (
	"fmt"
	"sync"

	"github.com/ptrimble/dreamhost-personal-backup/backup/logger"
)

type gatherResult struct {
	result map[string]File
	err    error
}

type processor struct {
	gatherLocalFiles  func() (map[string]File, error)
	gatherRemoteFiles func() (map[string]File, error)
	logger            logger.BackupLogger
	wg                *sync.WaitGroup
	remoteActions     chan<- RemoteAction
}

func NewProcessor(
	localGather, remoteGather func() (map[string]File, error),
	log logger.BackupLogger,
	wg *sync.WaitGroup,
	rac chan<- RemoteAction,
) processor {
	return processor{
		gatherLocalFiles:  localGather,
		gatherRemoteFiles: remoteGather,
		logger:            log,
		wg:                wg,
		remoteActions:     rac,
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
		p.logger.Error(logger.LogEntry{
			Message: fmt.Sprintf("error returned while gathering local files, err: %s", local.err),
		})

		return local.err
	}

	//FIXME If the remote is somehow wrong then this hangs!
	remote := <-remoteResultChan
	if remote.err != nil {
		p.logger.Error(logger.LogEntry{
			Message: fmt.Sprintf("error returned while gathering remote files, err: %s", remote.err),
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

func (p processor) processLocalVsRemote(local, remote map[string]File) {
	defer p.wg.Done()

	for lkey, lfile := range local {
		rfile, found := remote[lkey]
		if !found || !isEqual(lfile, rfile) {
			p.wg.Add(1)
			p.remoteActions <- RemoteAction{
				Type: PUSH,
				File: lfile,
			}
		}
	}
}

func (p processor) processRemoteVsLocal(local, remote map[string]File) {
	defer p.wg.Done()

	for rkey, rfile := range remote {
		_, found := local[rkey]
		if !found {
			p.wg.Add(1)
			p.remoteActions <- RemoteAction{
				Type: REMOVE,
				File: rfile,
			}
		}
	}
}
