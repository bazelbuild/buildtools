/*
Copyright 2024 Google LLC

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

// Package pipe contains a pipe that buffers input until it is started.
package pipe

import (
	"bytes"
	"io"
	"sync"
)

// Delayed is a pipe that buffers input until it is started.
type Delayed struct {
	buffer  bytes.Buffer
	writer  io.WriteCloser
	enabled bool
	mu      sync.Mutex
}

// New returns a delayed pipe.
func New(w io.WriteCloser) *Delayed {
	return &Delayed{
		writer: w,
	}
}

func (w *Delayed) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.enabled {
		return w.buffer.Write(p)
	}
	return w.writer.Write(p)
}

// Start flushes all buffered input in the pipe and causes subsequent input to be written through without buffering.
func (w *Delayed) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()
	io.Copy(w.writer, &w.buffer)
	w.enabled = true
}

// Close closes the pipe.
func (w *Delayed) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writer.Close()
}
