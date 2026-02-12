package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func newAssetsServer(next http.Handler) http.Handler {
	as := http.NewServeMux()
	as.HandleFunc("/thumbnails/{image}", thumbnailHandler)
	as.HandleFunc("/comics/{comicId}/pages/{page}", comicPageHandler)
	as.Handle("/", next)
	return as
}

func thumbnailHandler(w http.ResponseWriter, r *http.Request) {
	fname := r.PathValue("image")
	if getFileType(fname) != "image" {
		http.Error(w, "requested file is not an image", http.StatusBadRequest)
		return
	}

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
	pageN, err := strconv.Atoi(r.PathValue("page"))
	if err != nil || strings.TrimSpace(comid) == "" || pageN <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	streamComicPage(w, comid, pageN)
}
