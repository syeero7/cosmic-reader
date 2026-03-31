package main

import (
	"context"
	"errors"
	"io"

	"github.com/klauspost/compress/zip"
	"github.com/mholt/archives"
)

func convertToCBZ(id, fpath string, saveThumbnail func(string, io.Reader) error) (*Archive, error) {
	ex := new(Extractor)
	dst, cbz, err := ex.createCBZ(id)
	if err != nil {
		return nil, err
	}

	defer dst.Close()
	defer cbz.Close()

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

		defer func() { ex.counter += 1 }()
		if ex.counter == 0 {
			tr := io.TeeReader(tmpf, imgw)
			return saveThumbnail(id, tr)
		}

		_, err = io.Copy(imgw, tmpf)
		return err
	})

	ex.archive.PageCount = ex.counter
	return &ex.archive, err
}

func getComicPage(id string, page int) (*zip.File, error) {
	fpath, err := findArchive(id)
	if err != nil {
		return nil, err
	}

	cbz, err := zip.OpenReader(fpath)
	if err != nil {
		return nil, err
	}

	// close zip file
	idx := page - 1
	if idx < 0 || idx >= len(cbz.File) {
		return nil, errors.New("invalid page index")
	}

	file := cbz.File[idx]
	if getFileType(file.Name) != "image" {
		return nil, errors.New("file is not an image")
	}

	return file, nil
}

var tempComic = new(struct {
	pageCount int
	path      string
	pages     map[int]int
})

func extractTempComic(fpath string) (int, error) {
	cbz, err := zip.OpenReader(fpath)
	if err != nil {
		return 0, err
	}

	tempComic.path = fpath
	tempComic.pages = make(map[int]int, len(cbz.File))

	for i, file := range cbz.File {
		if getFileType(file.Name) != "image" {
			tempComic.pages[i+1] = -1
			continue
		}

		tempComic.pages[i+1] = i
		tempComic.pageCount++
	}

	return tempComic.pageCount, nil
}

func getTempComicInfo() int {
	return tempComic.pageCount
}

func getTempComicPage(page int) (*zip.File, error) {
	if tempComic == nil {
		return nil, errors.New("temporary comic not found")
	}

	cbz, err := zip.OpenReader(tempComic.path)
	if err != nil {
		return nil, err
	}

	idx, ok := tempComic.pages[page]
	if i := page - 1; !ok || i < 0 || i >= len(cbz.File) || idx == -1 {
		return nil, errors.New("invalid page index")
	}

	return cbz.File[idx], nil
}
