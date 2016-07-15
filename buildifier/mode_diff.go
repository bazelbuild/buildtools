package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	core "github.com/bazelbuild/buildifier/core"
	"github.com/bazelbuild/buildifier/differ"
)

func diffFiles(files []string) []error {
	// Start nworker workers reading stripes of the input
	// argument list and sending the resulting data on
	// separate channels. file[k] is read by worker k%nworker
	// and delivered on ch[k%nworker].
	type result struct {
		filename  string
		infile    string
		outfile   string
		tempfiles []string
		err       error
	}

	if len(files) == 0 {
		// nothing to process
		return nil
	}

	var wg sync.WaitGroup

	in := make(chan string)
	ch := make(chan result, len(files))

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			for filename := range in {
				equal, infile, outfile, tempfiles, err := writeDiffToTempFile(filename)
				if err != nil {
					ch <- result{filename: filename, err: err}
				}

				if equal {
					continue
				}

				ch <- result{filename, infile, outfile, tempfiles, err}
			}
			wg.Done()
		}()
	}
	for _, fname := range files {
		in <- fname
	}
	close(in)
	wg.Wait()
	close(ch)

	diff := differ.Find()

	var (
		errors    []error
		tempfiles []string
	)

	for res := range ch {
		if res.err != nil {
			errors = append(errors, res.err)
			continue
		}
		tempfiles = append(tempfiles, res.tempfiles...)
		diff.Show(res.infile, res.outfile)
	}

	diff.Run()

	// cleanup
	for _, fname := range tempfiles {
		os.Remove(fname)
	}

	return errors
}

func format(filename string) (in, out []byte, err error) {
	var f *core.File

	in, err = read(filename)
	f, err = core.Parse(filename, in)
	if err != nil {
		return nil, nil, err
	}

	core.Format(f)
	core.Rewrite(f)
	out = core.Format(f)
	return
}

// writeDiffToTempFile processes a single file containing data.
// It has been read from filename and should be written back if fixing.
func writeDiffToTempFile(filename string) (equal bool, infile, outfile string, tempfiles []string, err error) {
	var (
		data  []byte
		ndata []byte
	)

	equal = true
	infile = filename
	outfile = filename
	tempfiles = []string{}

	data, ndata, err = format(filename)
	if err != nil {
		return
	}

	if bytes.Equal(data, ndata) {
		return
	}

	equal = false

	outfile, err = writeTemp(ndata)
	if err != nil {
		return
	}
	tempfiles = append(tempfiles, outfile)

	if filename == "" || filename == "-" || filename == "stdin" {
		// data was read from standard filename.
		// Write it to a temporary file so diff can read it.
		infile, err = writeTemp(data)
		if err != nil {
			return
		}
		tempfiles = append(tempfiles, outfile)
	}

	return
}

// writeTemp writes data to a temporary file and returns the name of the file.
func writeTemp(data []byte) (file string, err error) {
	f, err := ioutil.TempFile("", "buildifier-tmp-")
	if err != nil {
		return "", fmt.Errorf("creating temporary file: %v", err)
	}
	name := f.Name()
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return "", fmt.Errorf("writing temporary file: %v", err)
	}
	return name, nil
}
