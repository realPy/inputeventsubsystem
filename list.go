package inputeventsubsystem

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func ScanInputs(inputpath string) []string {
	var devinputs []string

	if files, err := ioutil.ReadDir(inputpath); err == nil {

		for _, file := range files {
			var pathinput string = filepath.Join(inputpath, file.Name())
			if fileInfo, err := os.Lstat(pathinput); err == nil {

				if fileInfo.Mode()&os.ModeDevice == os.ModeDevice {

					devinputs = append(devinputs, pathinput)

				}

			}

		}
	}

	return devinputs
}
