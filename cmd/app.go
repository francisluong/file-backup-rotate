/*
Copyright © 2024 Francis Luong

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

var logger *log.Logger

func formatBackupFile(filePath string, backupNumID int) string {
	return fmt.Sprintf("%v.%v", filePath, backupNumID)
}

func doFirstBackup(fc *fileCopier) {
	fc.shouldCompareHash = true
	fc.CopyFile()
	if fc.err != nil {
		logger.Printf("Error returned: %v", fc.err)
	} else {
		logger.Printf("Completed - Last Action: %v", fc.actionDescr)
	}
}

func rotatePreviousBackups(filePath string, maxCount int) {
	// delete maxCount backup if it exists - ignore any errors
	// os.Remove(formatBackupFile(filePath, maxCount))
	var err error
	// ...then rename backups from maxcount-1 -> 1
	for i := maxCount - 1; i > 0; i-- {
		currFilePath := formatBackupFile(filePath, i)
		nextFilePath := formatBackupFile(filePath, i+1)
		err = os.Rename(currFilePath, nextFilePath)
		if err == nil {
			logger.Printf("rotated backup %v -> .%v", currFilePath, i+1)
		}
	}
}

func doBackup(filePath string, maxCount int) {
	backupFilePath := formatBackupFile(filePath, 1)
	logger.Printf("filePath: %v -> backupFilePath: %v", filePath, backupFilePath)
	// precheck whether copy/backup is indicated...
	fc := NewFileCopier(filePath, backupFilePath)
	if fc.PrecheckCopyIsNeeded() {
		// ...because we don't wan't to rotate files if backup is not indicated
		rotatePreviousBackups(filePath, maxCount)
		doFirstBackup(fc)
	} else {
		logger.Printf("Backup not needed: %v", fc.actionDescr)
	}
}

func ProcessFile() {
	filePath := viper.GetString("args.filePath")
	maxCount := viper.GetInt("args.maxCount")
	logger.Printf("ProcessFile ARGS: filePath: %v, maxCount: %v", filePath, maxCount)
	doBackup(filePath, maxCount)
}

func Init() {
	logger = log.New(os.Stdout, "file-backup-rotate: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("Init!")
}
