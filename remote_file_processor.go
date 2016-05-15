package backup

import (
	minio "github.com/minio/minio-go"
)

type remoteFileProcessor struct {
	listRemoteObjects func(string, string, bool, chan struct{}) <-chan minio.ObjectInfo
	fileData          map[string]file
}

func NewRemoteFileProcessor(lo func(string, string, bool, chan struct{}) <-chan minio.ObjectInfo) remoteFileProcessor {
	return remoteFileProcessor{
		listRemoteObjects: lo,
		fileData:          make(map[string]file, 0),
	}
}

func (p *remoteFileProcessor) Gather(t string) (data map[string]file, err error) {
	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	isRecursive := true
	objectCh := p.listRemoteObjects(t, "", isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		p.fileData[object.Key] = newFile(object.Key, object.Size)
	}

	return p.fileData, nil
}
