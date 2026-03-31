package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/disintegration/imaging"
)

var ThumbnailNotFoundError = errors.New("cached thumbnail not found")

type CacheDirName string

const (
	CacheThumbnails CacheDirName = "thumbnails"
	CacheAppState   CacheDirName = "state"
)

func getCachedThumbnail(fname string) (string, error) {
	cache, err := getCacheDir(CacheThumbnails)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(cache)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() || getFileType(entry.Name()) != "image" {
			continue
		}

		if entry.Name() == fname {
			return filepath.Join(cache, fname), nil
		}
	}

	return "", ThumbnailNotFoundError
}

func getCacheDir(dname CacheDirName) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(cache, "cosmic-reader", string(dname))
	err = os.MkdirAll(dir, 0700)
	return dir, err
}

func getHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(home, "cosmic-reader/archives")
	err = os.MkdirAll(dir, 0700)
	return dir, err
}

func getConfigDir() (string, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(config, "cosmic-reader")
	err = os.MkdirAll(dir, 0700)
	return dir, err
}

func createThumbnail(file io.Reader) ([]byte, error) {
	img, err := imaging.Decode(file)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	thumb := imaging.Resize(img, 0, 230, imaging.Linear)
	err = imaging.Encode(buf, thumb, imaging.JPEG, imaging.JPEGQuality(80))
	return buf.Bytes(), err
}

func deleteCachedThumbnails(fname *string) error {
	cache, err := getCacheDir(CacheThumbnails)
	if err != nil {
		return err
	}

	if fname != nil {
		return os.Remove(filepath.Join(cache, *fname))
	}

	files, err := filepath.Glob(filepath.Join(cache, "*"))
	if err != nil {
		return err
	}

	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			return err
		}
	}

	return nil
}
