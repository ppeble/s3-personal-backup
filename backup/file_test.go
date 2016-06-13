package backup

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FileEqual_Happy(t *testing.T) {
	f1 := File{
		Name: "file1",
		Size: 100,
	}

	f2 := File{
		Name: "file1",
		Size: 100,
	}

	assert.True(t, isEqual(f1, f2))
}

func Test_File_NotEqual_Name(t *testing.T) {
	f1 := File{
		Name: "file1",
		Size: 100,
	}

	f2 := File{
		Name: "file2",
		Size: 100,
	}

	assert.False(t, isEqual(f1, f2))
}

func Test_File_NotEqual_Size(t *testing.T) {
	f1 := File{
		Name: "file1",
		Size: 100,
	}

	f2 := File{
		Name: "file1",
		Size: 0,
	}

	assert.False(t, isEqual(f1, f2))
}

func Test_File_newFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "convertFromFileInfoDir")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)

	tmpFile, err := ioutil.TempFile(dir, "convertFromFileInfoFile")
	if err != nil {
		t.Fatal(err)
	}

	tempFileInfo, err := tmpFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	expected := File{
		Name: tmpFile.Name(),
		Size: tempFileInfo.Size(),
	}

	assert.Equal(t, expected, newFile(tmpFile.Name(), tempFileInfo.Size()))
}

func Test_File_String(t *testing.T) {
	f1 := File{
		Name: "file1",
		Size: 100,
	}

	assert.Equal(t, "name: 'file1' - size: '100'", f1.String())
}
