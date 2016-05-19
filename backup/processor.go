package backup

type processor struct {
	gatherLocalFiles  func() (map[string]file, error)
	gatherRemoteFiles func() (map[string]file, error)
	putToRemote       func(string) error
	removeFromRemote  func(string) error
}

func NewProcessor(
	localGather, remoteGather func() (map[string]file, error),
	putToRemote, removeFromRemote func(string) error,
) processor {
	return processor{
		gatherLocalFiles:  localGather,
		gatherRemoteFiles: remoteGather,
		putToRemote:       putToRemote,
		removeFromRemote:  removeFromRemote,
	}
}

func (p processor) Process() (err error) {
	localFiles, err := p.gatherLocalFiles()
	if err != nil {
		return
	}

	remoteFiles, err := p.gatherRemoteFiles()
	if err != nil {
		return
	}

	p.processLocalVsRemote(localFiles, remoteFiles)
	p.processRemoteVsLocal(localFiles, remoteFiles)

	return
}

func (p processor) processLocalVsRemote(local, remote map[string]file) {
	for lkey, lfile := range local {
		rfile, found := remote[lkey]
		if !found || !isEqual(lfile, rfile) {
			p.putToRemote(lkey)
		}
	}
}

func (p processor) processRemoteVsLocal(local, remote map[string]file) {
	for rkey, _ := range remote {
		_, found := local[rkey]
		if !found {
			p.removeFromRemote(rkey)
		}
	}
}
