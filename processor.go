package backup

// Make a constructor that accepts methods to call
// This way we can test this glue package
// Then in main we pass in real things

type processor struct {
	gatherLocalFiles  func(string) (map[string]file, error)
	gatherRemoteFiles func(string) (map[string]file, error)
	// etc
}

func NewProcessor(localGather, remoteGather func(string) (map[string]file, error)) processor {
	return processor{
		gatherLocalFiles:  localGather,
		gatherRemoteFiles: remoteGather,
	}
}

func (p processor) Process(targetLocal, targetRemote string) error {
	_, err := p.gatherLocalFiles(targetLocal)
	return err
}
