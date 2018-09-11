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

// ErrorCode is the API error code type.
type ErrorCode int

// API error codes.
const (
	// Factory errors
	IdInUse ErrorCode = iota
	InvalidIdFormat

	// Container errors
	ContainerNotExists
	ContainerPaused
	ContainerNotStopped
	ContainerNotRunning
	ContainerNotPaused

	// Process errors
	NoProcessOps

	// Common errors
	ConfigInvalid
	ConsoleExists
	SystemError
)

func (c ErrorCode) String() string {
	switch c {
	case IdInUse:
		return "Id already in use"
	case InvalidIdFormat:
		return "Invalid format"
	case ContainerPaused:
		return "Container paused"
	case ConfigInvalid:
		return "Invalid configuration"
	case SystemError:
		return "System error"
	case ContainerNotExists:
		return "Container does not exist"
	case ContainerNotStopped:
		return "Container is not stopped"
	case ContainerNotRunning:
		return "Container is not running"
	case ConsoleExists:
		return "Console exists for process"
	case ContainerNotPaused:
		return "Container is not paused"
	case NoProcessOps:
		return "No process operations"
	default:
		return "Unknown error"
	}
}

// Error is the API error type.
type Error interface {
	error

	// Returns a verbose string including the error message
	// and a representation of the stack trace suitable for
	// printing.
	Detail(w io.Writer) error

	// Returns the error code for this error.
	Code() ErrorCode
}
