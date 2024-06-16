package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBackupFile(t *testing.T) {
	assert.Equal(t, "save.dat.1", formatBackupFile("save.dat", 1))
	assert.Equal(t, "save.dat.2", formatBackupFile("save.dat", 2))
}
