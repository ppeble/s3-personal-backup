package backup

// These are the only two things that I am
// verifying for differences at this time. I am thinking
// about expanding it but for the time being this is enough.
type file struct {
	name string
	size int64
}

func isEqual(f1, f2 file) bool {
	return f1.name == f2.name &&
		f1.size == f2.size
}

func newFile(name string, size int64) file {
	return file{
		name: name,
		size: size,
	}
}
