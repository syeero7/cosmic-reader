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
	"strconv"
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

		filename := fmt.Sprintf("%0*d%s", padding, count, filepath.Ext(f.Name()))
		pagef, err := cbzw.Create(filename)
		if err != nil {
			return err
		}

		tmpf, err := f.Open()
		if err != nil {
			return err
		}

		defer tmpf.Close()
		if _, err := io.Copy(pagef, tmpf); err != nil {
			return err
		}

		count++
		return nil
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
	thumbnailDone, pageCountDone, titleDone := false, false, false

	err = extr.Extract(ctx, file, func(ctx context.Context, f archives.FileInfo) error {
		if strings.ToLower(f.Name()) == "comicinfo.xml" {
			pageCount, title, err := parseComicInfoXML(f)
			if err != nil && !errors.Is(err, PageCountParseError) {
				return err
			}

			cinfo.Title = title
			titleDone = true
			if pageCount > 0 {
				cinfo.PageCount = pageCount
				pageCountDone = true
			}
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

		if pageCountDone && thumbnailDone && titleDone {
			return ExtractionDoneError
		}

		return nil
	})

	if !pageCountDone {
		cinfo.PageCount = count
	}

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

func parseComicInfoXML(file archives.FileInfo) (int, string, error) {
	f, err := file.Open()
	if err != nil {
		return 0, "", err
	}

	defer f.Close()
	byt, err := io.ReadAll(f)
	if err != nil {
		return 0, "", err
	}

	metadata := new(ComicMetadata)
	if err := xml.Unmarshal(byt, metadata); err != nil {
		return 0, "", err
	}

	if strings.ToLower(metadata.Title) == "chapter" && len(metadata.Series) > 0 {
		metadata.Title = metadata.Series
	}

	if metadata.PageCount == 0 {
		parts := strings.Split(strings.ToLower(metadata.Summary), "pages: ")
		if len(parts) != 2 {
			return 0, metadata.Title, PageCountParseError
		}

		parts2 := strings.Fields(parts[1])
		if len(parts) == 0 {
			return 0, metadata.Title, PageCountParseError
		}

		n, err := strconv.Atoi(parts2[0])
		if err != nil {
			return 0, metadata.Title, PageCountParseError
		}

		metadata.PageCount = n
	}

	return metadata.PageCount, metadata.Title, nil
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
