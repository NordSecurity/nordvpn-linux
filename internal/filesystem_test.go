package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

const (
	TestDataPath       = "testdata/"
	TestFileReadFile   = "fread"
	TestFileDeleteFile = "fdelete"
	TestFileLockFile   = "flock"
	TestFileCopyFile   = "fcopy"
	TestFileSha256File = "fsha256"
	TestFileExistsFile = "fexists"
)

func TestManualListener(t *testing.T) {
	category.Set(t, category.Integration)

	listener, err := ManualListener("10601", PermUserRWGroupRWOthersRW, "pidfile")()
	assert.NoError(t, err)
	listener.Close()
}

func TestMachineID(t *testing.T) {
	category.Set(t, category.Integration)

	id := MachineID()

	hostname, err := os.Hostname()
	assert.NoError(t, err)

	// Test if MachineID is empty
	assert.NotEmpty(t, id.String())

	// Test if MachineID UUID contains hostname string
	assert.False(t, strings.Contains(id.String(), hostname))

	// Test if MachineID UUID contains hostname bytes
	byteStringSice := []byte(hostname)
	for index, hexVal := range byteStringSice {
		if !strings.Contains(id.String(), fmt.Sprintf("%x", int(hexVal))) {
			break
		}
		assert.NotEqual(t, index, len(byteStringSice)-1, "Machine ID contains hostname bytes")
	}
}

func TestEnsureDir(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []string{
		"singlefolderpath/filename",
		"multi/folder/path/filename",
		"filename",
	}

	for _, item := range tests {
		path := TestDataPath + item
		t.Run("ENSUREPATH="+item, func(t *testing.T) {
			err := EnsureDir(path)
			assert.Nil(t, err, "EnsureDir failed. Got error: %v", err)
			folders := strings.Split(path, "/")
			existsCheck := strings.Join(folders[:len(folders)-1], "/")
			if len(folders) > 2 {
				deletePath := strings.Join(folders[:2], "/")
				defer func() {
					os.RemoveAll(deletePath)
				}()
			}

			_, err = os.Stat(existsCheck)
			assert.NoError(t, err)
		})
	}
}

func TestFileExists(t *testing.T) {
	category.Set(t, category.File)

	tests := []struct {
		filePath string
		expected bool
	}{
		{TestFileExistsFile, true},
		{"fakefile", false},
		{"multi/folder/fake/file.txt", false},
		{".fakedotfile", false},
	}

	for _, item := range tests {
		got := FileExists(TestDataPath + item.filePath)
		assert.Equal(t, got, item.expected)
	}
}

func TestFileWrite(t *testing.T) {
	category.Set(t, category.File)

	tests := []struct {
		filename, data string
		permissions    os.FileMode
	}{
		{".d0tfile", "A quick brown fox jumps over a lazy dog", PermUserRWGroupROthersR},
		{"noext", "some extra text", PermUserRWGroupROthersR},
		{"withext.txt", "Lorem ipsum dolor sit amet, consectetur.", PermUserRWGroupROthersR},
	}

	for _, item := range tests {
		t.Run("FWNAME="+item.filename, func(t *testing.T) {
			path := TestDataPath + item.filename
			err := FileWrite(path, []byte(item.data), item.permissions)
			assert.NoError(t, err)

			file, err := os.OpenFile(path, os.O_RDONLY|os.O_EXCL, item.permissions)
			defer func() {
				file.Close()
				os.Remove(file.Name())
			}()

			got := make([]byte, len(item.data))
			_, err = file.Read(got)
			assert.NoError(t, err)
			assert.EqualValues(t, got, []byte(item.data))
		})
	}
}

func TestFileCreate(t *testing.T) {
	category.Set(t, category.File)

	tests := []struct {
		filename    string
		permissions os.FileMode
	}{
		{"withext.txt", PermUserRWGroupROthersR},
		{"noext", PermUserRWGroupROthersR},
		{".d0tfile", PermUserRWGroupROthersR},
		{"differentperm", PermUserRWGroupRWOthersRW},
	}

	for _, item := range tests {
		t.Run("FNAME="+item.filename, func(t *testing.T) {
			path := TestDataPath + item.filename
			file, err := FileCreate(path, item.permissions)
			assert.NoError(t, err)
			defer func() {
				file.Close()
				os.Remove(file.Name())
			}()
			_, filename := filepath.Split(file.Name())

			assert.Equal(t, filename, item.filename)

			stats, _ := file.Stat()
			assert.Equal(t, stats.Mode(), item.permissions)
		})
	}
}

func TestFileRead(t *testing.T) {
	category.Set(t, category.File)

	expected, err := ioutil.ReadFile(TestDataPath + TestFileReadFile)
	assert.NoError(t, err)
	got, err := FileRead(TestDataPath + TestFileReadFile)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, got)
}

func TestFileDelete(t *testing.T) {
	category.Set(t, category.File)

	filePath := TestDataPath + TestFileDeleteFile
	defer os.Remove(filePath)

	os.Create(filePath)
	err := FileDelete(filePath)
	assert.NoError(t, err)
	err = FileDelete(TestDataPath + "nonexistentfile")
	assert.Error(t, err)
}

func TestFileLock(t *testing.T) {
	category.Set(t, category.File)

	filePath := TestDataPath + TestFileLockFile
	os.Create(filePath)
	defer func() {
		FileUnlock(filePath)
		os.Remove(filePath)
	}()

	err := FileLock(filePath)
	assert.Nil(t, err)

	err = os.Remove(filePath)
	assert.Error(t, err, filePath)
}

func TestFileUnlock(t *testing.T) {
	category.Set(t, category.File)

	filePath := TestDataPath + TestFileLockFile
	os.Create(filePath)
	defer func() {
		exec.Command(ChattrExec, "-ia", filePath).Run()
		os.Remove(filePath)
	}()

	exec.Command(ChattrExec, "+i", filePath).Run()
	FileUnlock(filePath)
	err := os.Remove(filePath)
	assert.NoError(t, err)
}

func TestFileCopy(t *testing.T) {
	category.Set(t, category.File)

	filePath := TestDataPath + "fcopytemp"
	data, err := FileRead(TestDataPath + TestFileCopyFile)
	assert.NoError(t, err)
	err = FileCopy(TestDataPath+TestFileCopyFile, filePath)
	assert.NoError(t, err)
	defer FileDelete(filePath)
	copiedData, err := FileRead(filePath)
	assert.NoError(t, err)
	assert.EqualValues(t, data, copiedData)
}

func TestFileCopy_Invalid(t *testing.T) {
	category.Set(t, category.File)

	src := TestDataPath + "fdoesntexist"
	dst := TestDataPath + "fdoesntexisttemp"
	err := FileCopy(src, dst)
	defer os.Remove(dst)
	assert.Error(t, err)
}

func TestFileTemp(t *testing.T) {
	category.Set(t, category.File)

	sampleContent := []byte("Quick brown fox jumps over a lazy dog.")
	tempFilename := "ftemp"
	file, err := FileTemp(tempFilename, sampleContent)
	assert.NoError(t, err)
	defer func() {
		file.Close()
		FileDelete(file.Name())
	}()
	assert.Contains(t, file.Name(), tempFilename)
	got, err := FileRead(file.Name())
	assert.NoError(t, err)

	assert.EqualValues(t, sampleContent, got)
}

func TestFileSha256(t *testing.T) {
	category.Set(t, category.File)

	expected := "1F84CCB52684794248F981C2CC40B585C8106443244AC5197BB5601420246EAA"
	sha, err := FileSha256(TestDataPath + TestFileSha256File)
	assert.NoError(t, err)
	hexSha := fmt.Sprintf("%X", sha)
	assert.Equal(t, expected, hexSha)
	sha, err = FileSha256(TestDataPath + "fnonexistentsha256")
	assert.Error(t, err)
	assert.Nil(t, sha)
}

func TestIsCommandAvailable(t *testing.T) {
	category.Set(t, category.Integration)

	testData := []struct {
		command  string
		expected bool
	}{
		{"sh", true},
		{"cat", true},
		{"echo", true},
		{"fakecmd123", false},
		{"expre55vpn", false},
	}

	for _, item := range testData {
		got := IsCommandAvailable(item.command)
		assert.Equal(t, got, item.expected)
	}
}

func TestNetworkLinks(t *testing.T) {
	category.Set(t, category.Integration)

	ifaces, err := NetworkLinks()
	assert.NoError(t, err)

	for _, iface := range ifaces {
		assert.NotEmpty(t, iface.Name)
		regex := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$`)
		assert.True(t, regex.MatchString(iface.Address))
	}
}
