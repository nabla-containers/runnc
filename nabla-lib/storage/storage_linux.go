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

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lisofs
#define LIBISOFS_WITHOUT_LIBBURN yes
#include <stdint.h>
#include <libisofs/libisofs.h>
typedef struct burn_source struct_burn_source;

int burnSrc_read_xt(struct burn_source *burnSrc, void *buf, int len)
{
	return burnSrc->read_xt(burnSrc, buf, 2048);
}
*/
import "C"

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"unsafe"
	"path"
)

// CreateIso creates an ISO from the dir argument
func CreateIso(dir string) (string, error) {

	var image *C.IsoImage
	var opts *C.IsoWriteOpts
	var burnSrc *C.struct_burn_source
	var buf [2048]byte

	f, err := ioutil.TempFile("/tmp", "nabla")
	if err != nil {
		return "", err
	}

	defer f.Close()

	res := C.iso_init()
	if res < 0 {
		return "", fmt.Errorf("iso_init failed with: %v", res)
	}

	C.iso_set_msgs_severities(C.CString("NEVER"),
		C.CString("WARNING"), C.CString(""))
	res = C.iso_image_new(C.CString(""), &image)
	if res < 0 {
		return "", fmt.Errorf("iso_image_new failed with: %d", res)
	}

	C.iso_tree_set_follow_symlinks(image, 0)
	C.iso_tree_set_ignore_hidden(image, 1)
	C.iso_tree_set_ignore_special(image, 1)

	C.iso_set_abort_severity(C.CString("SORRY"))

	C.iso_set_local_charset(C.CString("UTF-8"), 1)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	C.iso_tree_add_exclude(image, C.CString(path.Join(absDir, "dev")));
	C.iso_tree_add_exclude(image, C.CString(path.Join(absDir, "sys")));
	C.iso_tree_add_exclude(image, C.CString(path.Join(absDir, "proc")));

	res = C.iso_tree_add_dir_rec(image,
		C.iso_image_get_root(image),
		C.CString(absDir))
	if res < 0 {
		return "",
			fmt.Errorf("iso_tree_add_dir_rec failed with: %d", res)
	}

	res = C.iso_write_opts_new(&opts, 1)
	if res < 0 {
		return "", fmt.Errorf("iso_write_opts_new failed with: %d", res)
	}

	res = C.iso_write_opts_set_rockridge(opts, 1)
	if res < 0 {
		return "", fmt.Errorf("set rockridge failed with: %d", res)
	}

	res = C.iso_image_create_burn_source(image, opts, &burnSrc)
	if res < 0 {
		return "", fmt.Errorf(
			"iso_image_create_burn_source failed with: %d, res")
	}

	for C.burnSrc_read_xt(burnSrc,
		unsafe.Pointer(&buf[0]), 2048) == 2048 {
		n, _ := f.Write(buf[:])
		if n < 2048 {
			break
		}
	}

	f.Sync()
	return f.Name(), nil
}

// CreateIso creates an ISO from the dir argument using
// the genisoimage command.
func CreateIsoCmd(dir string) (string, error) {
	f, err := ioutil.TempFile("/tmp", "nabla")
	if err != nil {
		return "", err
	}

	fname := f.Name()
	if err := f.Close(); err != nil {
		return "", err
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("genisoimage", "-m", "dev", "-m", "sys",
		"-m", "proc", "-r", "-o", fname, absDir)
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return fname, nil
}
