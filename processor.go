package backup

import (
	"fmt"
)

type processor struct {
	gatherLocalFiles  func() (map[string]file, error)
	gatherRemoteFiles func() (map[string]file, error)
}

func NewProcessor(localGather, remoteGather func() (map[string]file, error)) processor {
	return processor{
		gatherLocalFiles:  localGather,
		gatherRemoteFiles: remoteGather,
	}
}

func (p processor) Process() (err error) {
	localData, err := p.gatherLocalFiles()
	if err != nil {
		return
	}

	remoteData, err := p.gatherRemoteFiles()
	if err != nil {
		return
	}

	fmt.Printf("local: %#v\n", localData)
	fmt.Printf("remote: %#v\n", remoteData)

	return
}
