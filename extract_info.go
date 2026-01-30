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
	"strconv"
	"strings"

	"github.com/mholt/archives"
)

type ComicInfo struct {
	id            string
	pageCount     int
	title         string
	thumbnailPath string
}

type ComicMetadata struct {
	Title     string `xml:"Title"`
	Series    string `xml:"Series"`
	Number    string `xml:"Number"`
	Summary   string `xml:"Summary"`
	PageCount int    `xml:"PageCount"`
}

var PageCountParseError = errors.New("page count parsing failed")
var ExtractionDoneError = errors.New("info extraction is done")

func extractArchive(fpath string) (*ComicInfo, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	ctx := context.Background()
	format, _, err := archives.Identify(ctx, fpath, file)
	if err != nil {
		return nil, err
	}

	extr, ok := format.(archives.Extractor)
	if !ok || !isSupported(file.Name()) {
		return nil, fmt.Errorf("filetype '%s' is not supported", filepath.Ext(file.Name()))
	}

	count := 0
	var thumbnail archives.FileInfo
	cinfo := new(ComicInfo)
	thumbnailDone, pagesDone, titleDone := false, false, false
	err = extr.Extract(ctx, file, func(ctx context.Context, f archives.FileInfo) error {
		if strings.ToLower(f.NameInArchive) == "comicinfo.xml" {
			pageCount, title, err := parseComicInfoXML(f)
			if err != nil && !errors.Is(err, PageCountParseError) {
				return err
			}

			cinfo.title = title
			titleDone = true
			if pageCount > 0 {
				cinfo.pageCount = pageCount
				pagesDone = true
			}
		}

		if getFileType(f.NameInArchive) == "image" {
			n, err := getPageIndex(f.NameInArchive)
			if err != nil {
				return err
			}

			if n == 1 {
				thumbnail = f
				thumbnailDone = true
			}

			if titleDone && pagesDone && thumbnailDone {
				return ExtractionDoneError
			}

			count++
		}

		return nil
	})

	if !pagesDone {
		cinfo.pageCount = count
	}

	if !errors.Is(err, ExtractionDoneError) {
		return nil, err
	}

	id, err := getUniqueId(cinfo.title)
	if err != nil {
		return nil, err
	}

	cinfo.id = id
	// server thumbnail using http.ServeFile
	tmbpath, err := cacheThumbnail(thumbnail, id)
	cinfo.thumbnailPath = tmbpath
	return cinfo, err
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

func getPageIndex(name string) (int, error) {
	parts := strings.Split(name, ".")
	return strconv.Atoi(parts[0])
}
