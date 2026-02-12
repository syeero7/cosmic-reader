package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
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

// TODO: optimize page retrieving. optimize comic archive on adding for ez page retrieval

// archive length - 1 (comicInfo.xml) === total page count
// width := 6
// value := 12
// padded := fmt.Sprintf("%0*d", width, value)
// func countDigits(i int) int {
// 	if i == 0 {
// 		return 1
// 	}
//
// 	count := 0
// 	for i >= 0 {
// 		i /= 10
// 		count++
// 	}
// 	return count
// }

func streamComicPage(w http.ResponseWriter, id string, page int) error {
	fpath, err := storage.findArchive(id)
	if err != nil {
		return err
	}

	file, err := os.Open(fpath)
	if err != nil {
		return err
	}

	defer file.Close()
	ctx := context.Background()
	extr, err := getExtractor(file, ctx)
	if err != nil {
		return err
	}

	count := 0
	err = extr.Extract(ctx, file, func(ctx context.Context, f archives.FileInfo) error {
		if getFileType(f.Name()) == "image" {
			count++
			if count == page {
				img, err := f.Open()
				if err != nil {
					return err
				}

				defer img.Close()
				w.Header().Set("Cache-Control", "max-age=604800")
				w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(f.Name())))
				if _, err := io.Copy(w, img); err != nil {
					return err
				}

				return ExtractionDoneError
			}
		}

		return nil
	})

	if err != nil && !errors.Is(err, ExtractionDoneError) {
		return err
	}

	return nil
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
