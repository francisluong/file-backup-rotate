package cmd

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var licenseFilePath string = "../LICENSE"

func init() {
	logger = log.New(os.Stdout, "file-backup-rotate: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func TestDoFileSum(t *testing.T) {
	hash, err := DoFileSum(licenseFilePath)
	assert.Equal(t, "4002f795f7119311fc2413ef76e823dc38b3a59864c472c323c65089e1fd7861", hash)
	assert.Equal(t, nil, err)
	hash, err = DoFileSum("file-not-found")
	assert.Equal(t, "", hash)
	assert.EqualError(t, err, "open file-not-found: no such file or directory")
}

func TestCrudeBackup(t *testing.T) {
	destFilePath := "/tmp/TestCrudeBackup--LICENSE-copy.txt"
	defer os.Remove((destFilePath))
	CrudeBackup(licenseFilePath, destFilePath)
	hash, _ := DoFileSum(destFilePath)
	assert.Equal(t, "4002f795f7119311fc2413ef76e823dc38b3a59864c472c323c65089e1fd7861", hash)

}
