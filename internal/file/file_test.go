package file

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (m *MockFileInfo) Name() string {
	return m.name
}

func (m *MockFileInfo) Size() int64 {
	return m.size
}

func (m *MockFileInfo) Mode() os.FileMode {
	return m.mode
}

func (m *MockFileInfo) ModTime() time.Time {
	return m.modTime
}

func (m *MockFileInfo) IsDir() bool {
	return m.isDir
}

func (m *MockFileInfo) Sys() interface{} {
	return m.sys
}

func createFile(path string, data []byte) error {
	f, err := os.Create(path)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)

	if err != nil {
		return err
	}

	return nil
}

func Test(t *testing.T) { TestingT(t) }

type FileTestSuite struct{}

var _ = Suite(&FileTestSuite{})

func (*FileTestSuite) TestCopyFileCopiesFile(c *C) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	srcBytes := []byte{'t', 'e', 's', 't'}

	err := createFile(srcFile, srcBytes)
	c.Check(err, IsNil)
	defer os.Remove(srcFile)

	dstFile := filepath.Join(baseDir, "dst")
	err = CopyFile(srcFile, dstFile)
	c.Check(err, IsNil)
	defer os.Remove(dstFile)

	src, err := os.Open(srcFile)
	c.Check(err, IsNil)
	defer src.Close()
	var srcData []byte
	_, err = src.Read(srcData)
	c.Check(err, IsNil)

	dst, err := os.Open(dstFile)
	c.Check(err, IsNil)
	defer dst.Close()
	var dstData []byte
	_, err = dst.Read(dstData)
	c.Check(err, IsNil)

	c.Check(srcData, DeepEquals, dstData)
}

func (*FileTestSuite) TestCopyFileErrorsOnNonExistentSourceFile(c *C) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	dstFile := filepath.Join(baseDir, "dst")

	err := CopyFile(srcFile, dstFile)

	c.Check(err, Not(IsNil))
}

func (*FileTestSuite) TestCopyFileOverwritesExistingDst(c *C) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	srcBytes := []byte{'t', 'e', 's', 't'}

	err := createFile(srcFile, srcBytes)
	c.Check(err, IsNil)
	defer os.Remove(srcFile)

	dstFile := filepath.Join(baseDir, "dst")
	dstBytes := []byte{'t', 'e'}

	err = createFile(dstFile, dstBytes)
	c.Check(err, IsNil)
	defer os.Remove(dstFile)

	err = CopyFile(srcFile, dstFile)
	c.Check(err, IsNil)

	src, err := os.Open(srcFile)
	c.Check(err, IsNil)
	defer src.Close()
	var srcData []byte
	_, err = src.Read(srcData)
	c.Check(err, IsNil)

	dst, err := os.Open(dstFile)
	c.Check(err, IsNil)
	defer dst.Close()
	var dstData []byte
	_, err = dst.Read(dstData)
	c.Check(err, IsNil)

	c.Check(srcData, DeepEquals, dstData)
}

func (*FileTestSuite) TestHashFileCreatesCorrectHash(c *C) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	srcBytes := []byte{'t', 'e', 's', 't'}

	err := createFile(srcFile, srcBytes)
	c.Check(err, IsNil)
	defer os.Remove(srcFile)

	hashBytes := md5.Sum(srcBytes)
	hashString := hex.EncodeToString(hashBytes[:])

	testHashString, err := hashFile(srcFile)
	c.Check(err, IsNil)

	c.Check(hashString, Equals, testHashString)
}

func (*FileTestSuite) TestHashFileReturnsErrorIfFileDoesNotExist(c *C) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")

	_, err := hashFile(srcFile)
	c.Check(err, Not(IsNil))
}

func (*FileTestSuite) TestGenerateBackupDetailsAddsDirectoriesNotInBackupLocation(c *C) {
	srcIndex := map[string]os.FileInfo{
		"dir1": &MockFileInfo{
			name:    "dir1",
			size:    1,
			mode:    os.ModeDir,
			modTime: time.Now(),
			isDir:   true,
		},
		"dir2": &MockFileInfo{
			name:    "dir2",
			size:    1,
			mode:    os.ModeDir,
			modTime: time.Now(),
			isDir:   true,
		},
	}

	dstIndex := map[string]os.FileInfo{}

	srcDir := "/src/"
	dstDir := "/dst/"

	files, directories, symlinks := GenerateBackupDetails(srcIndex, dstIndex, srcDir, dstDir)

	c.Check(files, HasLen, 0)
	c.Check(symlinks, HasLen, 0)
	c.Check(directories, HasLen, 2)
	c.Check(directories, DeepEquals, []string{
		"dir1",
		"dir2",
	})
}

func (*FileTestSuite) TestGenerateBackupDetailsAddsFilesNotInBackupLocation(c *C) {
	srcIndex := map[string]os.FileInfo{
		"file1": &MockFileInfo{
			name:    "file1",
			size:    1,
			mode:    100644,
			modTime: time.Now(),
			isDir:   false,
		},
		"file2": &MockFileInfo{
			name:    "file2",
			size:    1,
			mode:    100644,
			modTime: time.Now(),
			isDir:   false,
		},
	}

	dstIndex := map[string]os.FileInfo{}

	srcDir := "/src/"
	dstDir := "/dst/"

	files, directories, symlinks := GenerateBackupDetails(srcIndex, dstIndex, srcDir, dstDir)

	c.Check(files, HasLen, 2)
	c.Check(symlinks, HasLen, 0)
	c.Check(directories, HasLen, 0)
	c.Check(files, DeepEquals, []string{
		"file1",
		"file2",
	})
}

func (*FileTestSuite) TestGenerateBackupDetailsAddsSymlinksNotInBackupLocation(c *C) {
	baseDir := os.TempDir()
	targetFile := filepath.Join(baseDir, "target")

	err := createFile(targetFile, []byte{})
	c.Check(err, IsNil)
	defer os.Remove(targetFile)

	symlink1 := filepath.Join(baseDir, "symlink1")
	err = os.Symlink(targetFile, symlink1)
	c.Check(err, IsNil)
	defer os.Remove(symlink1)

	symlink2 := filepath.Join(baseDir, "symlink2")
	err = os.Symlink(targetFile, symlink2)
	c.Check(err, IsNil)
	defer os.Remove(symlink2)

	srcIndex := map[string]os.FileInfo{
		"symlink1": &MockFileInfo{
			name:    "symlink1",
			size:    1,
			mode:    os.ModeSymlink,
			modTime: time.Now(),
			isDir:   false,
		},
		"symlink2": &MockFileInfo{
			name:    "symlink2",
			size:    1,
			mode:    os.ModeSymlink,
			modTime: time.Now(),
			isDir:   false,
		},
	}

	dstIndex := map[string]os.FileInfo{}

	srcDir := baseDir
	dstDir := "/dst/"

	files, directories, symlinks := GenerateBackupDetails(srcIndex, dstIndex, srcDir, dstDir)

	c.Check(files, HasLen, 0)
	c.Check(symlinks, HasLen, 2)
	c.Check(directories, HasLen, 0)
	c.Check(symlinks, DeepEquals, map[string]string{
		"symlink1": "target",
		"symlink2": "target",
	})
}
