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
	"slices"
	"strings"

	"github.com/klauspost/compress/zip"
	"github.com/mholt/archives"
)

type Extractor struct {
	archive Archive
}

type ComicMetadata struct {
	Title     string `xml:"Title"`
	Series    string `xml:"Series"`
	Number    string `xml:"Number"`
	Summary   string `xml:"Summary"`
	PageCount int    `xml:"PageCount"`
}

func (ex *Extractor) extract(fpath string, ctx context.Context, fn func(ctx context.Context, f archives.FileInfo) error) error {
	file, err := os.Open(fpath)
	if err != nil {
		return err
	}

	defer file.Close()
	extr, ext, err := ex.getExtractor(file, ctx)
	if err != nil {
		return err
	}

	if ext != ".cbz" {
		return extr.Extract(ctx, file, fn)
	}

	return nil
}

func (ex *Extractor) getExtractor(file *os.File, ctx context.Context) (archives.Extractor, string, error) {
	format, _, err := archives.Identify(ctx, file.Name(), file)
	if err != nil {
		return nil, "", err
	}

	ext := strings.ToLower(format.Extension())
	extr, ok := format.(archives.Extractor)
	if !ok || !isSupported(file.Name()) {
		return nil, "", fmt.Errorf("filetype '%s' is not supported", filepath.Ext(file.Name()))
	}

	return extr, ext, nil
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

func (ex *Extractor) extractComicTitle(xmlf fs.File, name, fpath string) error {
	title, err := parseTitleFromXML(xmlf, name)
	if err != nil {
		return err
	}

	if title == "" {
		title = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
	}

	ex.archive.Title = title
	return nil
}

func parseTitleFromXML(file fs.File, filename string) (string, error) {
	if strings.ToLower(filename) != "comicinfo.xml" {
		return "", nil
	}

	byt, err := io.ReadAll(file)
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
	fileTypes := []string{".cbr", ".cb7", ".cbt"}
	ext := strings.ToLower(filepath.Ext(fpath))
	return slices.Contains(fileTypes, ext)
}
