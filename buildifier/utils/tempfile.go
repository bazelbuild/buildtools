package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Tempfile creates temporary files and cleans them up in the end
type TempFile struct {
	filenames []string
}

// WriteTemp writes data to a temporary file and returns the name of the file.
func (tf *TempFile) WriteTemp(data []byte) (file string, err error) {
	f, err := ioutil.TempFile("", "buildifier-tmp-")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %v", err)
	}
	defer f.Close()
	name := f.Name()
	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("writing temporary file: %v", err)
	}
	tf.filenames = append(tf.filenames, name)
	return name, nil
}

// Clean removes all created temporary files
func (tf *TempFile) Clean() {
	for _, file := range tf.filenames {
		os.Remove(file)
	}
	tf.filenames = []string{}
}
