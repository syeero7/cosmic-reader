package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options"
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
	a.initializeDB()
	a.emitFileOpening(nil)
}

func (a *App) shutdown(ctx context.Context) {
	openedCBZ.reset()
	a.db.close()
}

func (a *App) assetServer(next http.Handler) http.Handler {
	a.initializeDB()
	return newAssetsServer(next, a.db)
}

func (a *App) secondInstanceLaunch(data options.SecondInstanceData) {
	a.emitFileOpening(data.Args)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
}

func (a *App) initializeDB() {
	if a.db != nil {
		return
	}

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

		if err := openedCBZ.extract(fpath); err != nil {
			log.Fatal(err)
		}

		runtime.EventsEmit(a.ctx, "comic-opened", openedCBZ.getInfo())
	}
}

func (a *App) GenerateULID() string {
	id, err := generateULID()
	if err != nil {
		log.Println(err)
	}

	return id.String()
}

func (a *App) SelectAnyComic() string {
	return a.selectFile(false)
}

func (a *App) SelectOnlyCBZ() string {
	return a.selectFile(true)
}

func (a *App) selectFile(onlyCBZ bool) string {
	opt := runtime.OpenDialogOptions{
		Title:                "Select Comic Books",
		CanCreateDirectories: false,
		Filters: []runtime.FileFilter{{
			DisplayName: "Comic Books (*.cbz, *.cbr, *.cbt, *.cb7)",
			Pattern:     "*.cbz;*.cbr;*.cb7;*.cbt",
		}},
	}

	if onlyCBZ {
		opt.Title = "Open a CBZ File"
		opt.Filters = []runtime.FileFilter{{
			DisplayName: "Comic Book Zip *.cbz",
			Pattern:     "*.cbz",
		}}
	}

	fpath, err := runtime.OpenFileDialog(a.ctx, opt)
	if err != nil {
		log.Fatal(err)
	}

	return fpath
}

func (a *App) DeleteComic(id string) {
	openedCBZ.reset()
	if err := a.db.removeArchive(id); err != nil {
		log.Fatal(err)
	}
}

func (a *App) AddComicBook(id, fpath string) string {
	openedCBZ.reset()
	arch, err := a.db.addArchive(id, fpath)
	if err != nil {
		log.Fatal(err)
	}

	return arch.Title
}

func (a *App) GetComicList() map[string]string {
	archives, err := a.db.getAllArchives()
	if err != nil {
		log.Fatal(err)
	}

	return archives
}

func (a *App) OpenCBZByID(id string) *ComicInfo {
	return a.openCBZFile(&id, nil)
}

func (a *App) OpenCBZByPath(fpath string) *ComicInfo {
	return a.openCBZFile(nil, &fpath)
}

func (a *App) openCBZFile(id, fpath *string) *ComicInfo {
	if id == nil && fpath == nil {
		return nil
	}

	if id != nil && fpath == nil {
		abspath, err := a.db.getAbsolutePath(*id)
		if err != nil {
			log.Fatal(err)
		}

		fpath = &abspath
	}

	if err := openedCBZ.extract(*fpath); err != nil {
		log.Fatal(err)
	}

	return openedCBZ.getInfo()
}

func (a *App) GetInitialOpenedCBZ() *ComicInfo {
	return openedCBZ.getInfo()
}
