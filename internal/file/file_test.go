package file

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
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

func TestCopyFileCopiesFile(t *testing.T) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	srcBytes := []byte{'t', 'e', 's', 't'}

	err := createFile(srcFile, srcBytes)
	if err != nil {
		t.Errorf("Error creating source file: %s", err)
	}
	defer os.Remove(srcFile)

	dstFile := filepath.Join(baseDir, "dst")
	err = CopyFile(srcFile, dstFile)
	if err != nil {
		t.Errorf("Error copying source file to destination: %s", err)
	}
	defer os.Remove(dstFile)

	src, err := os.Open(srcFile)
	if err != nil {
		t.Errorf("Error opening source file: %s", err)
	}
	defer src.Close()
	var srcData []byte
	_, err = src.Read(srcData)
	if err != nil {
		t.Errorf("Error reading source file: %s", err)
	}

	dst, err := os.Open(dstFile)
	if err != nil {
		t.Errorf("Error opening destination file: %s", err)
	}
	defer dst.Close()
	var dstData []byte
	_, err = dst.Read(dstData)
	if err != nil {
		t.Errorf("Error reading destination file: %s", err)
	}

	if bytes.Compare(srcData, dstData) != 0 {
		t.Errorf("Error: source and destination file contents not equal")
	}
}

func TestCopyFileErrorsOnNonExistentSourceFile(t *testing.T) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	dstFile := filepath.Join(baseDir, "dst")

	err := CopyFile(srcFile, dstFile)

	if err == nil {
		t.Errorf("CopyFile did not return an error")
	}
}

func TestCopyFileOverwritesExistingDst(t *testing.T) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	srcBytes := []byte{'t', 'e', 's', 't'}

	err := createFile(srcFile, srcBytes)
	if err != nil {
		t.Errorf("Error creating source file: %s", err)
	}
	defer os.Remove(srcFile)

	dstFile := filepath.Join(baseDir, "dst")
	dstBytes := []byte{'t', 'e'}

	err = createFile(dstFile, dstBytes)
	if err != nil {
		t.Errorf("Error creating destination file: %s", err)
	}
	defer os.Remove(dstFile)

	err = CopyFile(srcFile, dstFile)

	if err != nil {
		t.Errorf("Error copying source file to destination: %s", err)
	}

	src, err := os.Open(srcFile)
	if err != nil {
		t.Errorf("Error opening source file: %s", err)
	}
	defer src.Close()
	var srcData []byte
	_, err = src.Read(srcData)
	if err != nil {
		t.Errorf("Error reading source file: %s", err)
	}

	dst, err := os.Open(dstFile)
	if err != nil {
		t.Errorf("Error opening destination file: %s", err)
	}
	defer dst.Close()
	var dstData []byte
	_, err = dst.Read(dstData)
	if err != nil {
		t.Errorf("Error reading destination file: %s", err)
	}

	if bytes.Compare(srcData, dstData) != 0 {
		t.Errorf("Error: source and destination file contents not equal")
	}
}

func TestHashFileCreatesCorrectHash(t *testing.T) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")
	srcBytes := []byte{'t', 'e', 's', 't'}

	err := createFile(srcFile, srcBytes)
	if err != nil {
		t.Errorf("Error creating source file: %s", err)
	}
	defer os.Remove(srcFile)

	hashBytes := md5.Sum(srcBytes)
	hashString := hex.EncodeToString(hashBytes[:])

	testHashString, err := hashFile(srcFile)

	if err != nil {
		t.Errorf("hashFile returned an error: %s", err)
	}

	if hashString != testHashString {
		t.Errorf("Data hash %s not equal to file hash %s", hashString, testHashString)
	}
}

func TestHashFileReturnsErrorIfFileDoesNotExist(t *testing.T) {
	baseDir := os.TempDir()
	srcFile := filepath.Join(baseDir, "src")

	_, err := hashFile(srcFile)

	if err == nil {
		t.Errorf("hashFile did not return an error")
	}
}

func TestGenerateBackupDetailsAddsDirectoriesNotInBackupLocation(t *testing.T) {
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

	if len(files) != 0 {
		t.Errorf("List of files is not none")
	}

	if len(symlinks) != 0 {
		t.Errorf("List of symlinks is not none")
	}

	if len(directories) != 2 {
		t.Errorf("List of directories should be 2")
	}

	expectedDirs := []string{
		"dir1",
		"dir2",
	}

	if reflect.DeepEqual(directories, expectedDirs) {
		t.Errorf("directories contents does not match expected")
	}
}

func TestGenerateBackupDetailsAddsFilesNotInBackupLocation(t *testing.T) {
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

	if len(files) != 2 {
		t.Errorf("List of files should be sized 2")
	}

	expectedFiles := []string{
		"file1",
		"file2",
	}

	if reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("files contents does not match expected")
	}

	if len(symlinks) != 0 {
		t.Errorf("List of symlinks is not none")
	}

	if len(directories) != 0 {
		t.Errorf("List of directories is not none")
	}
}

func TestGenerateBackupDetailsAddsSymlinksNotInBackupLocation(t *testing.T) {
	baseDir := os.TempDir()
	targetFile := filepath.Join(baseDir, "target")

	err := createFile(targetFile, []byte{})
	if err != nil {
		t.Errorf("Error creating file: %s", err)
	}
	defer os.Remove(targetFile)

	symlink1 := filepath.Join(baseDir, "symlink1")
	err = os.Symlink(targetFile, symlink1)
	if err != nil {
		t.Errorf("Error creating file: %s", err)
	}
	defer os.Remove(symlink1)

	symlink2 := filepath.Join(baseDir, "symlink2")
	err = os.Symlink(targetFile, symlink2)
	if err != nil {
		t.Errorf("Error creating file: %s", err)
	}
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

	if len(symlinks) != 2 {
		t.Errorf("List of symlinks should be of size 2")
	}

	expectedSymlinks := map[string]string{
		"symlink1": targetFile,
		"symlink2": targetFile,
	}

	if reflect.DeepEqual(symlinks, expectedSymlinks) {
		t.Errorf("symlinks contents does not match expected")
	}

	if len(files) != 0 {
		t.Errorf("List of files is not none")
	}

	if len(directories) != 0 {
		t.Errorf("List of directories is not none")
	}
}
