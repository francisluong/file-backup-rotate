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
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"os"
)

type fileCopier struct {
	readPath          string
	readFD            *os.File
	readBuf           *bufio.Reader
	writePath         string
	writeFD           *os.File
	writeBuf          *bufio.Writer
	err               error
	actionDescr       string
	fileSumsMatch     bool
	shouldCompareHash bool
	verbose           bool
}

func NewFileCopier(filepath string, backupFilePath string) *fileCopier {
	fc := &fileCopier{
		readPath:          filepath,
		writePath:         backupFilePath,
		fileSumsMatch:     false,
		shouldCompareHash: false,
		verbose:           false,
		actionDescr:       "init",
	}
	return fc
}

func (fc *fileCopier) CopyFile() {
	fc.compareFileSums()
	fc.openReadFD()
	fc.openWriteFD()
	fc.doCopy()
	fc.tearDown()
}

func (fc *fileCopier) PrecheckCopyIsNeeded() bool {
	// return true if file copy should occur
	fc.shouldCompareHash = true
	fc.compareFileSums()
	return !fc.shouldNotContinue()
}

func (fc *fileCopier) shouldNotContinue() bool {
	if fc.err != nil || fc.fileSumsMatch {
		return true
	} else {
		return false
	}
}

func (fc *fileCopier) compareFileSums() {
	fc.actionDescr = "comparing file names"
	if fc.readPath == fc.writePath {
		fc.fileSumsMatch = true
		fc.actionDescr = "Confirmed: File Paths Match"
		return
	}
	// to pick up file not found errors, we calc read sum...
	fc.actionDescr = "calc readFile sum"
	readFileSum, readErr := DoFileSum(fc.readPath)
	if fc.verbose {
		logger.Printf("readFileSum: %v", readFileSum)
	}
	fc.err = readErr
	if fc.shouldNotContinue() || !fc.shouldCompareHash {
		return
	}
	fc.actionDescr = "calc writeFile sum"
	writeFileSum, _ := DoFileSum(fc.writePath)
	if fc.verbose {
		logger.Printf("writeFileSum: %v", writeFileSum)
	}
	if readFileSum == writeFileSum {
		fc.fileSumsMatch = true
	}
	fc.actionDescr = "Confirmed: File Sums Match"
}

func (fc *fileCopier) openReadFD() {
	if fc.shouldNotContinue() {
		return
	}
	fc.actionDescr = "open reader"
	fc.readFD, fc.err = os.Open(fc.readPath)
	logger.Printf("opened reader for %v", fc.readPath)
	if fc.shouldNotContinue() {
		return
	}
	fc.readBuf = bufio.NewReader(fc.readFD)
}

func (fc *fileCopier) openWriteFD() {
	if fc.shouldNotContinue() {
		return
	}
	fc.actionDescr = "open writer"
	fc.writeFD, fc.err = os.Create(fc.writePath)
	if fc.shouldNotContinue() {
		return
	}
	logger.Printf("opened writer for %v", fc.writePath)
	fc.writeBuf = bufio.NewWriter(fc.writeFD)
}

func (fc *fileCopier) doCopy() {
	buf := make([]byte, 1024)
	var bytesReadCount int
	if fc.verbose {
		logger.Print("init: doCopy")
	}
	for {
		if fc.shouldNotContinue() {
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
			if fc.shouldNotContinue() {
				return
			}
			// write a chunk
			fc.actionDescr = "loop: write buffer"
			_, fc.err = fc.writeBuf.Write(buf[:bytesReadCount])
			if fc.verbose {
				logger.Printf(" - buffered: %v", bytesReadCount)
			}
		}
	}
}

func (fc *fileCopier) tearDown() {
	if fc.shouldNotContinue() {
		logger.Printf("last action: %v", fc.actionDescr)
	}
	fc.readFD.Close()
	fc.writeFD.Close()
}

type FileHasher struct {
	readPath    string
	readFD      *os.File
	readBuf     *bufio.Reader
	hasher      hash.Hash
	writeBuf    *bufio.Writer
	err         error
	actionDescr string
}

func (fh *FileHasher) _initReader() {
	if fh.err != nil {
		return
	}
	fh.actionDescr = "open reader"
	fh.readFD, fh.err = os.Open(fh.readPath)
	logger.Printf("opened reader for FileHasher %v", fh.readPath)
}

func (fh *FileHasher) _initReadBuf() {
	if fh.err != nil {
		return
	}
	fh.actionDescr = "init read buffer"
	fh.readBuf = bufio.NewReader(fh.readFD)
}

func (fh *FileHasher) _initHasher() {
	if fh.err != nil {
		return
	}
	fh.actionDescr = "open hasher"
	fh.hasher = sha256.New()
}

func (fh *FileHasher) _initWriteBuf() {
	if fh.err != nil {
		return
	}
	fh.actionDescr = "init write buffer"
	fh.writeBuf =
		bufio.NewWriter(fh.hasher)
}

func (fh *FileHasher) GetSum() []byte {
	fh._initReader()
	defer fh.readFD.Close()
	fh._initReadBuf()
	fh._initHasher()
	fh._initWriteBuf()
	buf := make([]byte, 1024)
	var bytesReadCount int
	for {
		if fh.err != nil {
			return nil
		}
		// read a chunk
		fh.actionDescr = "loop: read file contents"
		bytesReadCount, fh.err = fh.readBuf.Read(buf)
		if bytesReadCount == 0 {
			// exit condition: zero bytes read and err will be EOF
			fh.actionDescr = "loop EXIT: flush buffer"
			if fh.err = fh.writeBuf.Flush(); fh.err != nil {
				return nil
			}
			fh.actionDescr = "loop EXIT: write successful!"
			return fh.hasher.Sum(nil)
		} else {
			if fh.err != nil {
				return nil
			}
			// write a chunk
			fh.actionDescr = "loop: write buffer"
			_, fh.err = fh.writeBuf.Write(buf[:bytesReadCount])
		}
	}
}

func DoFileSum(filepath string) (string, error) {
	fh := &FileHasher{readPath: filepath}
	return hex.EncodeToString(fh.GetSum()), fh.err
}
