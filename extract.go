package main

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archives"
)

// TODO: add cbz files without converting
func convertToCBZ(id, fpath string, saveThumbnail func(string, io.Reader) error) (*Archive, error) {
	ex := new(Extractor)
	dst, cbz, err := ex.createCBZ(id)
	if err != nil {
		return nil, err
	}

	defer dst.Close()
	defer cbz.Close()
	firstPage := true
	err = ex.extract(fpath, context.Background(), func(ctx context.Context, f archives.FileInfo) error {
		tmpf, err := f.Open()
		if err != nil {
			return err
		}

		defer tmpf.Close()
		if err := ex.extractComicTitle(tmpf, f.Name(), fpath); err != nil {
			return err
		}

		if getFileType(f.Name()) != "image" {
			return nil
		}

		imgw, err := ex.createZipEntry(cbz, tmpf)
		if err != nil {
			return err
		}

		if firstPage {
			firstPage = false
			tr := io.TeeReader(tmpf, imgw)
			return saveThumbnail(id, tr)
		}

		_, err = io.Copy(imgw, tmpf)
		return err
	})

	if err != nil {
		defer os.Remove(dst.Name())
	}

	ex.archive.Path = filepath.Base(dst.Name())
	return &ex.archive, err
}
