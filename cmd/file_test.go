package cmd

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var licenseFilePath string = "../LICENSE"
var destFilePath string = "/tmp/TestCrudeBackup--LICENSE-copy.txt"

func init() {
	logger = log.New(os.Stdout, "file-backup-rotate: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestDoFileSum(t *testing.T) {
	t.Parallel()
	hash, err := DoFileSum(licenseFilePath)
	assert.Equal(t, "4002f795f7119311fc2413ef76e823dc38b3a59864c472c323c65089e1fd7861", hash)
	assert.Equal(t, nil, err)
	hash, err = DoFileSum("file-not-found")
	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "open file-not-found: no such file or directory")
}

func TestFileCopier_NewFileCopier(t *testing.T) {
	t.Parallel()
	fc := NewFileCopier(licenseFilePath, destFilePath)
	assert.Equal(t, licenseFilePath, fc.readPath)
	assert.Equal(t, destFilePath, fc.writePath)
	assert.Equal(t, false, fc.fileSumsMatch)
	assert.Equal(t, false, fc.shouldCompareHash)
	assert.Equal(t, false, fc.verbose)
	assert.Equal(t, "init", fc.actionDescr)
}

func TestFileCopier_shouldNotContinue(t *testing.T) {
	t.Parallel()
	fc := NewFileCopier(licenseFilePath, destFilePath)
	// initially it shoudld not have a reason to stop
	assert.Equal(t, false, fc.fileSumsMatch)
	assert.Equal(t, nil, fc.err)
	assert.Equal(t, false, fc.shouldNotContinue())
	// should stop if fileSumsMatch is true
	fc = NewFileCopier(licenseFilePath, licenseFilePath)
	fc.verbose = true
	fc.shouldCompareHash = true
	fc.compareFileSums()
	assert.Equal(t, true, fc.fileSumsMatch)
	assert.Equal(t, true, fc.shouldNotContinue())
	// should stop if fc.err is not nil
	fc = NewFileCopier(licenseFilePath, destFilePath)
	fc.err = errors.New("1")
	assert.NotEqual(t, nil, fc.err)
	assert.Equal(t, true, fc.shouldNotContinue())
}

func TestFileCopier_compareFileSums(t *testing.T) {
	t.Parallel()
	fc := NewFileCopier(licenseFilePath, "otherPath")
	// should return immediately if !shouldCompareHash
	assert.Equal(t, false, fc.shouldCompareHash)
	// otherwise should stop copy action if files match
	fc = NewFileCopier(licenseFilePath, licenseFilePath)
	fc.verbose = true
	fc.shouldCompareHash = true
	fc.compareFileSums()
	assert.Equal(t, true, fc.fileSumsMatch)
	assert.Equal(t, true, fc.shouldNotContinue())
	assert.Equal(t, "Confirmed: File Paths Match", fc.actionDescr)
}

func TestFileCopier_CopyFile(t *testing.T) {
	defer os.Remove((destFilePath))
	fc := NewFileCopier(licenseFilePath, destFilePath)
	fc.CopyFile()
	hash, _ := DoFileSum(destFilePath)
	assert.Equal(t, "4002f795f7119311fc2413ef76e823dc38b3a59864c472c323c65089e1fd7861", hash)
	assert.Equal(t, "loop EXIT: write successful!", fc.actionDescr)
	fc2 := NewFileCopier(licenseFilePath, destFilePath)
	fc2.shouldCompareHash = true
	fc2.CopyFile()
	assert.Equal(t, "Confirmed: File Sums Match", fc2.actionDescr)
}
