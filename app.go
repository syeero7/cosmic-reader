package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	db  *Database
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.db.close()
}

func (a *App) initializeDB() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	a.db = db
}

func (a *App) emitFileOpening(args []string) {
	if args == nil {
		args = os.Args
	}

	if len(args) < 2 {
		return
	}

	if fpath := args[1]; strings.ToLower(filepath.Ext(fpath)) == ".cbz" {
		pages, err := extractTempComic(fpath)
		if err != nil {
			log.Fatal(err)
		}

		runtime.EventsEmit(a.ctx, "comic-opened", pages)
	}
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
	if err := a.db.removeArchive(id); err != nil {
		log.Fatal(err)
	}
}

func (a *App) AddComicBook(id, fpath string) *Archive {
	arch, err := a.db.addArchive(id, fpath)
	if err != nil {
		log.Fatal(err)
	}

	return arch
}

func (a *App) GetComicInfo() map[string]*Archive {
	archives, err := a.db.getAllArchives()
	if err != nil {
		log.Fatal(err)
	}

	return archives
}

func (a *App) OpenCBZFile() int {
	opt := runtime.OpenDialogOptions{
		Title:                "Open Comic Book Zip Archive",
		CanCreateDirectories: false,
		Filters: []runtime.FileFilter{{
			DisplayName: "Comic Book Zip Archive *.cbz",
			Pattern:     "*.cbz",
		}},
	}

	fpath, err := runtime.OpenFileDialog(a.ctx, opt)
	if err != nil {
		log.Fatal(err)
	}

	if fpath == "" {
		return 0
	}

	pages, err := extractTempComic(fpath)
	if err != nil {
		log.Fatal(err)
	}

	return pages
}

func (a *App) GetInitialOpenedCBZ() int {
	return getTempComicInfo()
}
