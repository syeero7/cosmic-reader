package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"

	"github.com/disintegration/imaging"
)

var ThumbnailNotFoundError = errors.New("cached thumbnail not found")

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
