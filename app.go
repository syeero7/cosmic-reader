package main

import (
	"context"
	"log"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectFile() string {
	opt := runtime.OpenDialogOptions{
		Title:                "Select Comic Books",
		CanCreateDirectories: false,
		Filters: []runtime.FileFilter{{
			DisplayName: "Comic Books (*.cbz, *.cbr, *.cb7)",
			Pattern:     "*.cbz;*.cbr;*.cb7;*.cbt",
		}},
	}

	fpath, err := runtime.OpenFileDialog(a.ctx, opt)
	if err != nil {
		log.Fatal(err)
	}

	return fpath
}

func (a *App) DeleteComic(id string) {
	if err := storage.removeArchive(id); err != nil {
		log.Fatal(err)
	}
}

func (a *App) AddComicBook(id, fpath string) *Archive {
	arch, err := extractComicInfo(fpath, id)
	if err != nil {
		log.Fatal(err)
	}

	if err := storage.addArchive(id, fpath, *arch, false); err != nil {
		log.Fatal(err)
	}

	arch.Thumbnail = filepath.Base(arch.Thumbnail)
	return arch
}

type ArchiveInfo struct {
	Archive
	ID string `json:"id"`
}

func (a *App) GetComicInfo() []ArchiveInfo {
	state, err := storage.getState()
	if err != nil {
		log.Fatal(err)
	}

	archives := make([]ArchiveInfo, 0, len(state.Archives))
	for k, v := range state.Archives {
		v.Thumbnail = filepath.Base(v.Thumbnail)
		archives = append(archives, ArchiveInfo{Archive: v, ID: k})
	}

	return archives
}
