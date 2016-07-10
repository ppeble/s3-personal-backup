package backup

import (
	"fmt"
)

type Filename string

// These are the only two things that I am
// verifying for differences at this time. I am thinking
// about expanding it but for the time being this is enough.
type File struct {
	Name string
	Size int64
}

func newFile(name string, size int64) File {
	return File{
		Name: name,
		Size: size,
	}
}

func (f File) String() string {
	return fmt.Sprintf("name: '%s' - size: '%d'", f.Name, f.Size)
}

func (f File) Equal(otherFile File) bool {
	return f.Name == otherFile.Name &&
		f.Size == otherFile.Size
}
