package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type MainService struct {
	app *application.App
	db  *Database
}

func NewMainService(app *application.App) *MainService {
	return &MainService{app: app}
}

func (s *MainService) ServiceStartup(_ context.Context, options application.ServiceOptions) error {
	s.initializeDB()
	s.emitFileOpening(nil)
	return nil
}

func (s *MainService) ServiceShutdown() error {
	openedCBZ.reset()
	return nil
}

// FIXME: emit instance startup data
func (s *MainService) secondInstanceLaunch(data application.SecondInstanceData) {
	s.emitFileOpening(data.Args)
}

func (s *MainService) initializeDB() {
	if s.db != nil {
		return
	}

	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	s.db = db
}

func (s *MainService) emitFileOpening(args []string) {
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

		s.app.Event.Emit("comic-opened", openedCBZ.getInfo())
	}
}

func (s *MainService) GenerateULID() string {
	id, err := generateULID()
	if err != nil {
		log.Println(err)
	}

	return id.String()
}

func (s *MainService) SelectAnyComic() string {
	return s.selectFile(false)
}

func (s *MainService) SelectOnlyCBZ() string {
	return s.selectFile(true)
}

func (s *MainService) selectFile(onlyCBZ bool) string {
	opt := &application.OpenFileDialogOptions{
		Title:                "Select Comic Books",
		CanCreateDirectories: false,
		Filters: []application.FileFilter{{
			DisplayName: "Comic Books (*.cbz, *.cbr, *.cbt, *.cb7)",
			Pattern:     "*.cbz;*.cbr;*.cb7;*.cbt",
		}},
	}

	if onlyCBZ {
		opt.Title = "Open a CBZ File"
		opt.Filters = []application.FileFilter{{
			DisplayName: "Comic Book Zip *.cbz",
			Pattern:     "*.cbz",
		}}
	}

	fpath, err := s.app.Dialog.OpenFileWithOptions(opt).PromptForSingleSelection()
	if err != nil {

	}

	return fpath
}

func (s *MainService) DeleteComic(id string) {
	openedCBZ.reset()
	if err := s.db.removeArchive(id); err != nil {
		log.Fatal(err)
	}
}

func (s *MainService) AddComicBook(id, fpath string) string {
	openedCBZ.reset()
	arch, err := s.db.addArchive(id, fpath)
	if err != nil {
		log.Fatal(err)
	}

	return arch.Title
}

func (s *MainService) GetComicList() map[string]string {
	archives, err := s.db.getAllArchives()
	if err != nil {
		log.Fatal(err)
	}

	return archives
}

func (s *MainService) OpenCBZByID(id string) *ComicInfo {
	return s.openCBZFile(&id, nil)
}

func (s *MainService) OpenCBZByPath(fpath string) *ComicInfo {
	return s.openCBZFile(nil, &fpath)
}

func (s *MainService) openCBZFile(id, fpath *string) *ComicInfo {
	if id == nil && fpath == nil {
		return nil
	}

	if id != nil && fpath == nil {
		abspath, err := s.db.getAbsolutePath(*id)
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

func (s *MainService) GetInitialOpenedCBZ() *ComicInfo {
	return openedCBZ.getInfo()
}
