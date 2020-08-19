/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

// TempFile keeps track of temporary files and cleans them up in the end
type TempFile struct {
	filenames []string
}

// WriteTemp writes data to a temporary file and returns the name of the file.
func (tf *TempFile) WriteTemp(data []byte) (file string, err error) {
	f, err := ioutil.TempFile("", "buildifier-tmp-")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %v", err)
	}
	defer func() {
		e := f.Close()
		if e != nil {
			err = e
		}
	}()

	name := f.Name()
	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("writing temporary file: %v", err)
	}
	tf.filenames = append(tf.filenames, name)
	return name, nil
}

// Clean removes all created temporary files
func (tf *TempFile) Clean() error {
	for _, file := range tf.filenames {
		if err := os.Remove(file); err != nil {
			return err
		}
	}
	tf.filenames = []string{}
	return nil
}
