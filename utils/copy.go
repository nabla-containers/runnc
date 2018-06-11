package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// copyPath adapted from
// https://gist.github.com/elazarl/5507969 and
// https://github.com/otiai10/copy/blob/master/copy.go
// Changed to have different semantics and behaviors for file perms, overwrites
// and copy attributes
func Copy(dst, src string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return pcopy(dst, src, srcInfo)
}

func pcopy(dst, src string, srcInfo os.FileInfo) error {
	if srcInfo.IsDir() {
		return dcopy(dst, src, srcInfo)
	} else {
		return fcopy(dst, src, srcInfo)
	}
}

func fcopy(dst, src string, srcInfo os.FileInfo) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		// If file already exist, overwrite it
		if os.IsExist(err) {
			if d, err = os.OpenFile(dst, os.O_RDWR|os.O_TRUNC, 0755); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}

	err = os.Chmod(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	return d.Close()
}

func dcopy(dst, src string, srcInfo os.FileInfo) error {
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	infos, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if err := pcopy(
			filepath.Join(dst, info.Name()),
			filepath.Join(src, info.Name()),
			info,
		); err != nil {
			return err
		}
	}
	return nil

}
