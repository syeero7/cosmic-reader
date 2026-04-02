package main

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
)

func newAssetsServer(next http.Handler, db *Database) http.Handler {
	as := http.NewServeMux()
	as.HandleFunc("/thumbnails/{image}", databaseMiddleware(db, thumbnailHandler))
	as.HandleFunc("/pages/{page}", comicPageHandler)
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

func comicPageHandler(w http.ResponseWriter, r *http.Request) {
	pageN, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || pageN <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	file, err := openedCBZ.getComicPage(pageN)
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
