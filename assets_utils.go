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

func getCachedThumbnail(fname string) (string, error) {
	cache, err := getCacheDir()
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

func getCacheDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(cache, "cosmic-reader/thumbnails")
	err = os.MkdirAll(dir, 0700)
	return dir, err
}

func cacheThumbnail(finfo archives.FileInfo, fname string) (string, error) {
	file, err := finfo.Open()
	if err != nil {
		return "", err
	}

	img, err := imaging.Decode(file)
	if err != nil {
		return "", err
	}

	thumb := imaging.Thumbnail(img, 100, 150, imaging.CatmullRom)
	tmbname := fname + "." + imaging.JPEG.String()
	if err := imaging.Save(thumb, tmbname, imaging.JPEGQuality(80)); err != nil {
		return "", err
	}

	cache, err := getCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cache, tmbname), nil
}

func deleteCachedThumbnails(fname *string) error {
	cache, err := getCacheDir()
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
