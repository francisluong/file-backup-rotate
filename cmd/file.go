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
	"bufio"
	"os"
)

func CrudeBackup(filepath string, backupFilePath string) error {
	// sourced heavily from https://stackoverflow.com/a/9739903
	var err error
	var readFD *os.File
	var writeFD *os.File
	readFD, err = os.Open(filepath)
	if err != nil {
		return err
	} else {
		defer readFD.Close()
		logger.Printf("opened reader for %v", filepath)
	}
	readBufIO := bufio.NewReader(readFD)
	if err != nil {
		return err
	}

	writeFD, err = os.Create(backupFilePath)
	if err != nil {
		return err
	} else {
		logger.Printf("opened writer for %v", backupFilePath)
	}
	writeBufIO := bufio.NewWriter(writeFD)
	buf := make([]byte, 1024)
	var bytesReadCount int
	for {
		// read a chunk
		bytesReadCount, err = readBufIO.Read(buf)
		// exit condition: zero bytes read
		if bytesReadCount == 0 {
			break
		}
		// write a chunk
		writeBufIO.Write(buf[:bytesReadCount])
	}
	if bytesReadCount == 0 {
		if err := writeBufIO.Flush(); err != nil {
			return err
		} else {
			writeFD.Close()
			logger.Printf("Successfully wrote %v", backupFilePath)
			return nil
		}
	} else {
		return err
	}
}
