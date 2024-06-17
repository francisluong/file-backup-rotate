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

type fileCopier struct {
	readPath    string
	readFD      *os.File
	readBuf     *bufio.Reader
	writePath   string
	writeFD     *os.File
	writeBuf    *bufio.Writer
	err         error
	actionDescr string
}

func (fc *fileCopier) openReadFD() {
	if fc.err != nil {
		return
	}
	fc.actionDescr = "open reader"
	fc.readFD, fc.err = os.Open(fc.readPath)
	logger.Printf("opened reader for %v", fc.readPath)
	if fc.err != nil {
		return
	}
	fc.readBuf = bufio.NewReader(fc.readFD)
}

func (fc *fileCopier) openWriteFD() {
	if fc.err != nil {
		return
	}
	fc.actionDescr = "open writer"
	fc.writeFD, fc.err = os.Create(fc.writePath)
	if fc.err != nil {
		return
	}
	logger.Printf("opened writer for %v", fc.writePath)
	fc.writeBuf = bufio.NewWriter(fc.writeFD)
}

func (fc *fileCopier) doCopy(verbose bool) {
	buf := make([]byte, 1024)
	var bytesReadCount int
	if verbose {
		logger.Print("init: doCopy")
	}
	for {
		if fc.err != nil {
			return
		}
		// read a chunk
		fc.actionDescr = "loop: read file contents"
		bytesReadCount, fc.err = fc.readBuf.Read(buf)
		if bytesReadCount == 0 {
			// exit condition: zero bytes read and err will be EOF
			fc.actionDescr = "loop EXIT: flush buffer"
			if fc.err = fc.writeBuf.Flush(); fc.err != nil {
				return
			}
			fc.writeFD.Close()
			fc.actionDescr = "loop EXIT: write successful!"
			return
		} else {
			if fc.err != nil {
				return
			}
			// write a chunk
			fc.actionDescr = "loop: write buffer"
			_, fc.err = fc.writeBuf.Write(buf[:bytesReadCount])
			if verbose {
				logger.Printf(" - buffered: %v", bytesReadCount)
			}
		}
	}
}

func (fc *fileCopier) tearDown() {
	if fc.err != nil {
		logger.Printf("last action: %v", fc.actionDescr)
	}
	fc.readFD.Close()
	fc.writeFD.Close()
}

func CrudeBackup(filepath string, backupFilePath string) error {
	// sourced heavily from https://stackoverflow.com/a/9739903
	//    ...and https://go.dev/blog/errors-are-values
	fc := &fileCopier{readPath: filepath, writePath: backupFilePath}
	fc.openReadFD()
	fc.openWriteFD()
	fc.doCopy(false)
	fc.tearDown()
	return fc.err
}
