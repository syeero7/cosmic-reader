package main

import (
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/mholt/archives"
)

var CachedThumbnailNotFoundError = errors.New("cached thumbnail not found")

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

	return "", CachedThumbnailNotFoundError
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

func cacheThumbnail(finfo *archives.FileInfo, fname string) (string, error) {
	file, err := finfo.Open()
	if err != nil {
		return "", err
	}

	defer file.Close()
	cache, err := getCacheDir(CacheThumbnails)
	if err != nil {
		return "", err
	}

	tmbpath := filepath.Join(cache, fname+".jpeg")
	img, err := imaging.Decode(file)
	if err != nil {
		return "", err
	}

	thumb := imaging.Resize(img, 0, 230, imaging.Linear)
	err = imaging.Save(thumb, tmbpath, imaging.JPEGQuality(80))
	return tmbpath, err
}

func deleteCachedThumbnails(fname *string) error {
	cache, err := getCacheDir(CacheThumbnails)
	if err != nil {
		return err
	}

	if fname != nil {
		return os.Remove(filepath.Join(cache, filepath.Base(*fname)))
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
