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

func newAssetsServer(next http.Handler) http.Handler {
	as := http.NewServeMux()
	as.HandleFunc("/thumbnails/{image}", thumbnailHandler)
	as.HandleFunc("/comics/{comicId}/pages/{page}", comicPageHandler)
	as.Handle("/", next)
	return as
}

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
	fname := r.PathValue("image") + ".jpeg"
	img, err := getCachedThumbnail(fname)
	if err != nil {
		if errors.Is(err, CachedThumbnailNotFoundError) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, img)
}

func comicPageHandler(w http.ResponseWriter, r *http.Request) {
	comid := r.PathValue("comicId")
	tempComic := r.URL.Query().Get("temp") == "true"
	pageN, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || strings.TrimSpace(comid) == "" || pageN <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var file *zip.File
	if tempComic {
		file, err = getTempComicPage(pageN)
	} else {
		file, err = getComicPage(comid, pageN)
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
