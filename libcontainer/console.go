// Copyright 2014 Docker, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package libcontainer

import "io"

// Console represents a pseudo TTY.
type Console interface {
	io.ReadWriter
	io.Closer

	// Path returns the filesystem path to the slave side of the pty.
	Path() string

	// Fd returns the fd for the master of the pty.
	Fd() uintptr
}
