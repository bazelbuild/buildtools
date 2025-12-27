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

// Package color contains ANSI color codes.
package color

const keyEscape = 27

var (
	red    = []byte{keyEscape, '[', '3', '1', 'm'}
	green  = []byte{keyEscape, '[', '3', '2', 'm'}
	yellow = []byte{keyEscape, '[', '3', '3', 'm'}

	reset = []byte{keyEscape, '[', '0', 'm'}
)

func wrap(s string, codes []byte) string {
	return string(codes) + s + string(reset)
}

// Green returns s wrapped in ANSI codes which cause terminals to display it green.
func Green(s string) string {
	return wrap(s, green)
}

// Yellow returns s wrapped in ANSI codes which cause terminals to display it yellow.
func Yellow(s string) string {
	return wrap(s, yellow)
}

// Red returns s wrapped in ANSI codes which cause terminals to display it red.
func Red(s string) string {
	return wrap(s, red)
}
