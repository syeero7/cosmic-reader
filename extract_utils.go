package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zip"
	"github.com/mholt/archives"
)

type Extractor struct {
	archive Archive
	counter int
}

type ComicMetadata struct {
	Title     string `xml:"Title"`
	Series    string `xml:"Series"`
	Number    string `xml:"Number"`
	Summary   string `xml:"Summary"`
	PageCount int    `xml:"PageCount"`
}

func (ex *Extractor) getExtractor(file *os.File, ctx context.Context) (archives.Extractor, error) {
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

func (ex *Extractor) createCBZ(filename string) (*os.File, *zip.Writer, error) {
	home, err := getHomeDir()
	if err != nil {
		return nil, nil, err
	}

	dst, err := os.Create(filepath.Join(home, filename+".cbz"))
	if err != nil {
		return nil, nil, err
	}

	return dst, zip.NewWriter(dst), nil
}

func (ex *Extractor) createZipEntry(cbz *zip.Writer, file fs.File) (io.Writer, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return nil, err
	}

	if ext := strings.ToLower(filepath.Ext(stat.Name())); ext == ".bmp" || ext == ".raw" {
		header.Method = zip.Deflate
	} else {
		header.Method = zip.Store
	}

	writer, err := cbz.CreateHeader(header)
	if err != nil {
		return nil, err
	}

	return writer, err
}

func (ex *Extractor) extractComicTitle(xmlf *archives.FileInfo, filename string) error {
	title, err := parseTitleFromXML(xmlf)
	if err != nil {
		return err
	}

	if title == "" {
		title = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	}

	ex.archive.Title = title
	return nil
}

func (ex *Extractor) extractThumbnail(file io.Reader, id string) error {
	tmbpath, err := cacheThumbnail(file, id)
	if err != nil {
		return err
	}

	ex.archive.Thumbnail = filepath.Base(tmbpath)
	return nil
}

func parseTitleFromXML(file *archives.FileInfo) (string, error) {
	if strings.ToLower(file.Name()) != "comicinfo.xml" {
		return "", nil
	}

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
