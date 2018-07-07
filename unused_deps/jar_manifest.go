/*
Copyright 2018 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"strings"
)

// manifestIndexNewline scans manifest and returns the index of the first
// newline and the index of the byte following the newline.
func manifestIndexNewline(manifest string) (int, int, error) {
	// A newline in a jar manifest is denoted with CR LF, CR, or LF.
	n := strings.IndexByte(manifest, '\n')
	r := strings.IndexByte(manifest, '\r')
	if n < 0 && r < 0 {
		return -1, -1, fmt.Errorf("no newline in '%s'", manifest)
	}

	if n < 0 {
		// Only CR.
		return r, r + 1, nil
	}
	if r < 0 || n < r {
		// Only LF or we have both but the LF comes first.
		return n, n + 1, nil
	}

	// We have both CR and LF and the CR comes before the LF.  Check for the
	// special case of adjacent CR LF, which together denode a newline.
	if n == r+1 {
		return r, n + 1, nil
	}

	return r, r + 1, nil
}

// jarManifestValue returns the value associated with key in the manifest of
// the jar file named jarFileName.
func jarManifestValue(jarFileName string, key string) (string, error) {
	r, err := zip.OpenReader(jarFileName)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != "META-INF/MANIFEST.MF" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		bytes, err := ioutil.ReadAll(rc)
		if err != nil {
			return "", err
		}
		contents := string(bytes)

		found := strings.Index(contents, key+":")
		if found < 0 {
			return "", fmt.Errorf("manifest of jar %s contains no key %s", jarFileName, key)
		}

		label := ""
		rest := contents[found+len(key+":"):]
		for len(rest) > 0 && rest[0] == ' ' {
			newline, next, err := manifestIndexNewline(rest)
			if err != nil {
				return "", fmt.Errorf("bad value for key %s in manifest of jar %s", key, jarFileName)
			}
			label += rest[1:newline]
			rest = rest[next:]
		}
		return label, nil
	}
	return "", fmt.Errorf("jar file %s has no manifest", jarFileName)
}
