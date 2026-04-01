package main

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/klauspost/compress/zip"
)

func newAssetsServer(next http.Handler, db *Database) http.Handler {
	as := http.NewServeMux()
	as.HandleFunc("/thumbnails/{image}", databaseMiddleware(db, thumbnailHandler))
	as.HandleFunc("/comics/{comicId}/pages/{page}", databaseMiddleware(db, comicPageHandler))
	as.Handle("/", next)
	return as
}

func thumbnailHandler(db *Database, w http.ResponseWriter, r *http.Request) {
	img, err := db.getThumbnail(r.PathValue("image"))
	if err != nil {
		if errors.Is(err, ThumbnailNotFoundError) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "max-age=172800")
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(img)
}

func comicPageHandler(db *Database, w http.ResponseWriter, r *http.Request) {
	comid := r.PathValue("comicId")
	pageN, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || strings.TrimSpace(comid) == "" || pageN <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var file *zip.File
	if r.URL.Query().Get("temp") == "true" {
		file, err = getTempComicPage(pageN)
	} else {
		fpath, err := db.getAbsolutePath(comid)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		file, err = getComicPage(fpath, pageN)
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	img, err := file.Open()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	defer img.Close()
	w.Header().Set("Cache-Control", "max-age=172800")
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(file.Name)))
	io.Copy(w, img)
}

func databaseMiddleware(db *Database, fn func(db *Database, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(db, w, r)
	}
}
