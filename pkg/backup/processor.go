package backup

import (
	"fmt"
	"sync"
)

type processor struct {
	localGatherers []FileGatherer
	remoteGatherer FileGatherer
	logger         backupLogger
	wg             *sync.WaitGroup
	remoteActions  chan<- RemoteAction
}

func NewProcessor(
	localGatherers []FileGatherer,
	remoteGatherer FileGatherer,
	log backupLogger,
	wg *sync.WaitGroup,
	rac chan<- RemoteAction,
) processor {
	return processor{
		localGatherers: localGatherers,
		remoteGatherer: remoteGatherer,
		logger:         log,
		wg:             wg,
		remoteActions:  rac,
	}
}

func (p processor) Process() (err error) {
	//TODO should I do both gathering in separate goroutines here, like I do with the section below?
	localFiles, remoteFiles, err := p.runGatherers()
	if err != nil {
		return err
	}

	p.wg.Add(2)
	go p.processLocalVsRemote(localFiles, remoteFiles)
	go p.processRemoteVsLocal(localFiles, remoteFiles)

	return
}

func (p processor) runGatherers() (localFiles, remoteFiles FileData, err error) {
	//TODO These should be run in parallel. Only return when both are done without errors
	localFiles, err = p.runLocalGatherers(p.localGatherers)
	if err != nil {
		p.logger.Error(LogEntry{
			Message: fmt.Sprintf("error returned while gathering local files, err: %s", err),
		})

		return
	}

	remoteFiles, err = p.remoteGatherer.Gather()
	if err != nil {
		p.logger.Error(LogEntry{
			Message: fmt.Sprintf("error returned while gathering remote files, err: %s", err),
		})

		return
	}

	return
}

func (p processor) runLocalGatherers(localGatherers []FileGatherer) (FileData, error) {
	results := make([]map[Filename]File, 0)

	for _, g := range localGatherers {
		localFiles, err := g.Gather()
		if err != nil {
			return nil, err
		}

		results = append(results, localFiles)
	}

	combinedResults := make(FileData)
	for _, r := range results {
		for k, v := range r {
			combinedResults[k] = v
		}
	}

	return combinedResults, nil
}

func (p processor) processLocalVsRemote(local, remote FileData) {
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

func (p processor) processRemoteVsLocal(local, remote FileData) {
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
