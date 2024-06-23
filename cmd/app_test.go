package cmd

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testfile string = "root.go"
var testglob string = "root.go*"

func writeFile(filepath string, contents string) {
	writeFD, err := os.Create(filepath)
	if err != nil {
		log.Fatal(err)
	}
	writeFD.WriteString(contents)
	writeFD.Close()
}

func TestFormatBackupFile(t *testing.T) {
	assert.Equal(t, "save.dat.1", formatBackupFile("save.dat", 1))
	assert.Equal(t, "save.dat.2", formatBackupFile("save.dat", 2))
}

func TestRotatePreviousBackups_noPrev(t *testing.T) {
	rotatePreviousBackups(testfile, 2)
	matches, err := filepath.Glob(testglob)
	var expectedMatches []string
	expectedMatches = append(expectedMatches, "root.go")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedMatches, matches)
}

func TestRotatePreviousBackups_1_replaces_2(t *testing.T) {
	writeFile("root.go.1", "1")
	writeFile("root.go.2", "2")
	defer os.Remove("root.go.1")
	defer os.Remove("root.go.2")
	rotatePreviousBackups(testfile, 2)
	matches, err := filepath.Glob(testglob)
	var expectedMatches []string
	// .1 gets renamed to .2
	expectedMatches = append(expectedMatches, "root.go")
	expectedMatches = append(expectedMatches, "root.go.2")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedMatches, matches)
}

func TestRotatePreviousBackups_should_not_delete_maxCount(t *testing.T) {
	writeFile("root.go.2", "2")
	defer os.Remove("root.go.2")
	rotatePreviousBackups(testfile, 2)
	matches, err := filepath.Glob(testglob)
	var expectedMatches []string
	// because 2 == maxCount, but it is the only backup,
	//   therefore, it doesn't get deleted
	expectedMatches = append(expectedMatches, "root.go")
	expectedMatches = append(expectedMatches, "root.go.2")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedMatches, matches)
}

func TestDoBackup_no_previous_backups(t *testing.T) {
	defer os.Remove("root.go.1")
	defer os.Remove("root.go.2")
	doBackup(testfile, 2)
	matches, err := filepath.Glob(testglob)
	var expectedMatches []string
	expectedMatches = append(expectedMatches, "root.go")
	expectedMatches = append(expectedMatches, "root.go.1")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedMatches, matches)
	doBackup(testfile, 2)
	// on subsequent run, no new backups because files match
	matches2, err2 := filepath.Glob(testglob)
	assert.Equal(t, nil, err2)
	assert.Equal(t, expectedMatches, matches2)
}

func TestDoBackup_2_prev_backups(t *testing.T) {
	file1 := "root.go.1"
	file2 := "root.go.2"
	writeFile(file1, "1")
	writeFile(file2, "2")
	defer os.Remove(file1)
	defer os.Remove(file2)
	doBackup(testfile, 2)
	matches, err := filepath.Glob(testglob)
	var expectedMatches []string
	expectedMatches = append(expectedMatches, "root.go")
	expectedMatches = append(expectedMatches, "root.go.1")
	expectedMatches = append(expectedMatches, "root.go.2")
	assert.Equal(t, nil, err)
	assert.Equal(t, expectedMatches, matches)
	// run the backup
	doBackup(testfile, 2)
	// and a no-op run
	doBackup(testfile, 2)
	var filepaths []string
	var filesums []string
	filepaths = append(filepaths, testfile)
	filepaths = append(filepaths, file1)
	filepaths = append(filepaths, file2)
	for i := 0; i < len(filepaths); i++ {
		sum, _ := DoFileSum(filepaths[i])
		filesums = append(filesums, sum)
	}
	var expectedSums []string
	expectedSums = append(expectedSums, "84e4b309e5e1122657c3e26297ffe59e8e666494cfeb8bc7bc8db9b3d8403123")
	expectedSums = append(expectedSums, "84e4b309e5e1122657c3e26297ffe59e8e666494cfeb8bc7bc8db9b3d8403123")
	expectedSums = append(expectedSums, "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b")
	assert.Equal(t, expectedSums, filesums)
}
