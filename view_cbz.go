package main

import (
	"errors"
	"log"

	"github.com/klauspost/compress/zip"
)

type OpenedCBZ struct {
	pageCount int
	filepath  string
	pages     map[int]int
	cbz       *zip.ReadCloser
}

type ComicInfo struct {
	PageCount int    `json:"pageCount"`
	Title     string `json:"title"`
	ID        string `json:"id"`
}

var openedCBZ = new(OpenedCBZ)

func (o *OpenedCBZ) extract(fpath string) error {
	o.reset()
	cbz, err := zip.OpenReader(fpath)
	if err != nil {
		return err
	}

	o.cbz = cbz
	o.filepath = fpath
	o.pages = make(map[int]int, len(cbz.File))

	for i, file := range cbz.File {
		if getFileType(file.Name) != "image" {
			o.pages[i+1] = -1
			continue
		}

		o.pages[i+1] = i
		o.pageCount++
	}

	return nil
}

func (o *OpenedCBZ) getComicPage(n int) (*zip.File, error) {
	if o.cbz == nil {
		return nil, errors.New("no comic book currently open")
	}

	idx, ok := o.pages[n]
	if i := n - 1; !ok || i < 0 || i >= len(o.cbz.File) || idx == -1 {
		return nil, errors.New("invalid page index")
	}

	return o.cbz.File[idx], nil
}

func (o *OpenedCBZ) reset() {
	if o.cbz == nil {
		return
	}

	o.pageCount = 0
	if err := o.cbz.Close(); err != nil {
		log.Println(err)
	}

}

func (o *OpenedCBZ) getInfo() *ComicInfo {
	return &ComicInfo{PageCount: o.pageCount}
}
