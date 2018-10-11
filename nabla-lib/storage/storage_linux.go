// Copyright (c) 2018, IBM
// Author(s): Brandon Lum, Ricardo Koller
//
// Permission to use, copy, modify, and/or distribute this software for
// any purpose with or without fee is hereby granted, provided that the
// above copyright notice and this permission notice appear in all
// copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
// WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
// AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
// DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA
// OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// +build linux

package storage

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

// CreateDummy creates a dummy file in /tmp
func CreateDummy() (string, error) {
	file, err := ioutil.TempFile("/tmp", "nabla")
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

// CreateIso creates an ISO from the dir argument
func CreateIso(dir string, target *string) (string, error) {
	var fname string

	if target == nil {
		f, err := ioutil.TempFile("/tmp", "nabla")
		if err != nil {
			return "", err
		}

		fname = f.Name()
		if err := f.Close(); err != nil {
			return "", err
		}
	} else {
		var err error
		fname, err = filepath.Abs(*target)
		if err != nil {
			return "", errors.Wrap(err, "Unable to resolve abs target path")
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", errors.Wrap(err, "Unable to resolve abs dir path")
	}

	cmd := exec.Command("genisoimage", "-m", "dev", "-m", "sys",
		"-m", "proc", "-l", "-r", "-o", fname, absDir)
	err = cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "Unable to run geniso command")
	}

	return fname, nil
}
