package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

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

func extractComicInfo(fpath, id string) (*Archive, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	ctx := context.Background()
	extr, err := getExtractor(file, ctx)
	if err != nil {
		return nil, err
	}

	count := 0
	cinfo := new(Archive)
	var thumbnail archives.FileInfo
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

		if getFileType(f.Name()) == "image" {
			count++
			if count == 1 {
				thumbnail = f
				thumbnailDone = true
			}

			if pageCountDone && thumbnailDone && titleDone {
				return ExtractionDoneError
			}
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

	tmbpath, err := cacheThumbnail(thumbnail, id)
	cinfo.Thumbnail = tmbpath
	return cinfo, err
}

func extractComicPages(fpath string, pages []int) ([][]byte, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	ctx := context.Background()
	extr, err := getExtractor(file, ctx)
	if err != nil {
		return nil, err
	}

	count := 0
	comicPages := make([]archives.FileInfo, 0, len(pages))
	err = extr.Extract(ctx, file, func(ctx context.Context, f archives.FileInfo) error {
		if getFileType(f.Name()) == "image" {
			count++
			if slices.Contains(pages, count) {
				comicPages = append(comicPages, f)
			}

			if len(pages) == len(comicPages) {
				return ExtractionDoneError
			}
		}

		return nil
	})

	if !errors.Is(err, ExtractionDoneError) {
		return nil, err
	}

	cpages := make([][]byte, 0, len(pages))
	for _, page := range comicPages {
		byt, err := func() ([]byte, error) {
			f, err := page.Open()
			if err != nil {
				return []byte{}, err
			}

			defer f.Close()
			bt, err := io.ReadAll(f)
			return bt, err
		}()

		if err != nil {
			return nil, err
		}

		cpages = append(cpages, byt)
	}

	return cpages, nil
}

func getExtractor(file *os.File, ctx context.Context) (archives.Extractor, error) {
	format, _, err := archives.Identify(ctx, file.Name(), file)
	if err != nil {
		return nil, err
	}

	extr, ok := format.(archives.Extractor)
	if !ok || !isSupported(file.Name()) {
		return nil, fmt.Errorf("filetype '%s' is not supported", filepath.Ext(file.Name()))
	}

	return extr, nil
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

	// TODO: remove metadata logging
	slog.Any("comic metadata", metadata)
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
