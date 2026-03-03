package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zip"
	"github.com/mholt/archives"
)

type ComicMetadata struct {
	Title     string `xml:"Title"`
	Series    string `xml:"Series"`
	Number    string `xml:"Number"`
	Summary   string `xml:"Summary"`
	PageCount int    `xml:"PageCount"`
}

var PageCountParseError = errors.New("page count parsing failed")
var ExtractionDoneError = errors.New("info extraction is done")

func convertToCBZ(id, fpath string, pageCount int) error {
	file, extr, ctx, err := getExtractor(fpath)
	if err != nil {
		return err
	}

	defer file.Close()
	home, err := getHomeDir()
	if err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(home, id+".cbz"))
	if err != nil {
		return err
	}

	defer dst.Close()
	cbzw := zip.NewWriter(dst)
	defer cbzw.Close()

	count := 1
	padding := countDigits(pageCount) + 1
	err = extr.Extract(ctx, file, func(ctx context.Context, f archives.FileInfo) error {
		if getFileType(f.Name()) != "image" {
			return nil
		}

		filename := fmt.Sprintf("%0*d%s", padding, count, strings.ToLower(filepath.Ext(f.Name())))
		count++
		return addPageToCBZ(cbzw, f, filename)
	})

	return err
}

func extractComicInfo(id, fpath string) (*Archive, error) {
	file, extr, ctx, err := getExtractor(fpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	count := 0
	cinfo := new(Archive)
	thumbnailDone, titleDone := false, false

	err = extr.Extract(ctx, file, func(ctx context.Context, f archives.FileInfo) error {
		if strings.ToLower(f.Name()) == "comicinfo.xml" {
			title, err := parseTitleFromXML(&f)
			if err != nil && !errors.Is(err, PageCountParseError) {
				return err
			}

			cinfo.Title = title
			titleDone = true
		}

		if getFileType(f.Name()) != "image" {
			return nil
		}

		count++
		if count == 1 {
			tmbpath, err := cacheThumbnail(&f, id)
			if err != nil {
				return err
			}

			cinfo.Thumbnail = tmbpath
			thumbnailDone = true
		}

		if thumbnailDone && titleDone {
			return ExtractionDoneError
		}

		return nil
	})

	if cinfo.Title == "" {
		cinfo.Title = strings.TrimSuffix(filepath.Base(file.Name()), filepath.Ext(file.Name()))
	}

	if err != nil && !errors.Is(err, ExtractionDoneError) {
		return nil, err
	}

	return cinfo, err
}

func getComicPage(id string, page int) (*zip.File, error) {
	fpath, err := storage.findArchive(id)
	if err != nil {
		return nil, err
	}

	cbz, err := zip.OpenReader(fpath)
	if err != nil {
		return nil, err
	}

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
	path  string
	pages map[int]int
})

func extractTempComic(fpath string) (int, error) {
	cbz, err := zip.OpenReader(fpath)
	if err != nil {
		return 0, err
	}

	tempComic.path = fpath
	tempComic.pages = make(map[int]int, len(cbz.File))

	pageCount := 0
	for i, file := range cbz.File {
		if getFileType(file.Name) != "image" {
			tempComic.pages[i+1] = -1
			continue
		}

		tempComic.pages[i+1] = i
		pageCount++
	}

	return pageCount, nil
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
	if i := page - 1; i < 0 || i >= len(cbz.File) || !ok || idx == -1 {
		return nil, errors.New("invalid page index")
	}

	return cbz.File[idx], nil
}

func getExtractor(fpath string) (*os.File, archives.Extractor, context.Context, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx := context.Background()
	format, _, err := archives.Identify(ctx, file.Name(), file)
	if err != nil {
		return nil, nil, nil, err
	}

	extr, ok := format.(archives.Extractor)
	if !ok || !isSupported(file.Name()) {
		return nil, nil, nil, fmt.Errorf("filetype '%s' is not supported", filepath.Ext(file.Name()))
	}

	return file, extr, ctx, nil
}

func addPageToCBZ(cbz *zip.Writer, file archives.FileInfo, filename string) error {
	tmpf, err := file.Open()
	if err != nil {
		return err
	}

	defer tmpf.Close()
	stat, err := tmpf.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return err
	}

	if ext := filepath.Ext(filename); ext == ".bmp" || ext == ".raw" {
		header.Method = zip.Deflate
	} else {
		header.Method = zip.Store
	}

	header.Name = filename
	imgw, err := cbz.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(imgw, tmpf)
	return err
}

func parseTitleFromXML(file *archives.FileInfo) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}

	defer f.Close()
	byt, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	metadata := new(ComicMetadata)
	if err := xml.Unmarshal(byt, metadata); err != nil {
		return "", err
	}

	if strings.ToLower(metadata.Title) == "chapter" && len(metadata.Series) > 0 {
		metadata.Title = metadata.Series
	}

	return metadata.Title, nil
}

func getFileType(name string) string {
	mimetype := mime.TypeByExtension(filepath.Ext(name))
	return strings.Split(mimetype, "/")[0]
}

func isSupported(fpath string) bool {
	fileTypes := []string{"cbr", "cbz", "cb7", "cbt"}
	for _, t := range fileTypes {
		if strings.HasSuffix(fpath, t) {
			return true
		}
	}
	return false
}

func countDigits(n int) int {
	count := 0
	for n != 0 {
		n /= 10
		count++
	}

	return count
}
