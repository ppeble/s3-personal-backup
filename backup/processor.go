package backup

import (
	"fmt"
	"sync"

	"github.com/ptrimble/dreamhost-personal-backup/backup/logger"
)

type processor struct {
	localGathers      []func() (map[string]File, error)
	gatherRemoteFiles func() (map[string]File, error)
	logger            logger.BackupLogger
	wg                *sync.WaitGroup
	remoteActions     chan<- RemoteAction
}

func NewProcessor(
	localGathers []func() (map[string]File, error),
	remoteGather func() (map[string]File, error),
	log logger.BackupLogger,
	wg *sync.WaitGroup,
	rac chan<- RemoteAction,
) processor {
	return processor{
		localGathers:      localGathers,
		gatherRemoteFiles: remoteGather,
		logger:            log,
		wg:                wg,
		remoteActions:     rac,
	}
}

func (p processor) Process() (err error) {
	localFiles, err := p.runLocalGathers(p.localGathers)
	if err != nil {
		p.logger.Error(logger.LogEntry{
			Message: fmt.Sprintf("error returned while gathering local files, err: %s", err),
		})

		return err
	}

	remoteFiles, err := p.gatherRemoteFiles()
	if err != nil {
		p.logger.Error(logger.LogEntry{
			Message: fmt.Sprintf("error returned while gathering remote files, err: %s", err),
		})

		return err
	}

	p.wg.Add(2)
	go p.processLocalVsRemote(localFiles, remoteFiles)
	go p.processRemoteVsLocal(localFiles, remoteFiles)

	return
}

func (p processor) runLocalGathers(localGathers []func() (map[string]File, error)) (map[string]File, error) {
	results := make([]map[string]File, 0)

	for _, localGather := range localGathers {
		localFiles, err := localGather()
		if err != nil {
			return nil, err
		}

		results = append(results, localFiles)
	}

	combinedResults := make(map[string]File)
	for _, r := range results {
		for k, v := range r {
			combinedResults[k] = v
		}
	}

	return combinedResults, nil
}

func (p processor) processLocalVsRemote(local, remote map[string]File) {
	defer p.wg.Done()

	for lkey, lfile := range local {
		rfile, found := remote[lkey]
		if !found || !lfile.Equal(rfile) {
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
