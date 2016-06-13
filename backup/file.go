package backup

import (
	"fmt"
)

// These are the only two things that I am
// verifying for differences at this time. I am thinking
// about expanding it but for the time being this is enough.
type File struct {
	Name string
	Size int64
}

func (f File) String() string {
	return fmt.Sprintf("name: '%s' - size: '%d'", f.Name, f.Size)
}

func newFile(name string, size int64) File {
	return File{
		Name: name,
		Size: size,
	}
}

//TODO could this be a method off of the file struct?
// For example: f1.Equal(f2)?
func isEqual(f1, f2 File) bool {
	return f1.Name == f2.Name &&
		f1.Size == f2.Size
}
