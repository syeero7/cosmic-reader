package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type AppState struct {
	Settings struct {
		LibraryDir string `json:"libraryDir"`
	} `json:"settings"`
	Archives map[string]Archive `json:"archives"`
}

type Archive struct {
	PageCount int    `json:"pageCount"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
}

type StateManager struct {
	mu    sync.Mutex
	state *AppState
}

var storage StateManager

func (s *StateManager) setLibraryDir(dpath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.getState()
	if err != nil {
		return err
	}

	state.Settings.LibraryDir = dpath
	return s.save(state)
}

func (s *StateManager) getArchive(id string) (*Archive, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.getState()
	if err != nil {
		return nil, err
	}

	archive, ok := state.Archives[id]
	if !ok {
		return nil, fmt.Errorf("archive: %s not found", id)
	}

	return &archive, nil
}

func (s *StateManager) removeArchive(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.getState()
	if err != nil {
		return err
	}

	arch, ok := state.Archives[id]
	if !ok {
		return fmt.Errorf("archive: %s not found", id)
	}

	fpath, err := s.findArchive(id)
	if err != nil {
		return err
	}

	if err := os.Remove(fpath); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := deleteCachedThumbnails(&arch.Thumbnail); err != nil && !os.IsNotExist(err) {
		return err
	}

	delete(state.Archives, id)
	return s.save(state)
}

func (s *StateManager) addArchive(id, fpath string, replace bool) (*Archive, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.getState()
	if err != nil {
		return nil, err
	}

	_, ok := state.Archives[id]
	if !replace && ok {
		return nil, fmt.Errorf("archive: %s exists", id)
	}

	if replace && !ok {
		return nil, fmt.Errorf("archive: %s not found", id)
	}

	arch, err := extractComic(id, fpath)
	if err != nil {
		return nil, err
	}

	state.Archives[id] = *arch
	return arch, s.save(state)
}

func (s *StateManager) save(state *AppState) error {
	data, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}

	cache, err := getCacheDir(CacheAppState)
	if err != nil {
		return err
	}

	fpath := filepath.Join(cache, "state.json")
	tmp := fpath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	tmpf, err := os.Open(tmp)
	if err != nil {
		return err
	}

	tmpf.Sync()
	tmpf.Close()
	return os.Rename(tmp, fpath)
}

func (s *StateManager) getState() (*AppState, error) {
	if s.state != nil {
		return s.state, nil
	}

	cache, err := getCacheDir(CacheAppState)
	if err != nil {
		return nil, err
	}

	var state AppState
	fpath := filepath.Join(cache, "state.json")
	file, err := os.Open(fpath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil && file != nil {
		defer file.Close()
		if err := json.NewDecoder(file).Decode(&state); err != nil {
			return nil, err
		}
	}

	if os.IsNotExist(err) && (file == nil || state.Archives == nil) {
		return s.getDefaultState()
	}

	return &state, nil
}

func (s *StateManager) getDefaultState() (*AppState, error) {
	home, err := getHomeDir()
	if err != nil {
		return nil, err
	}

	state := new(AppState)
	state.Settings.LibraryDir = home
	state.Archives = make(map[string]Archive)
	return state, nil
}

var ArchiveFoundError = errors.New("archive found")

func (s *StateManager) findArchive(id string) (string, error) {
	home, err := getHomeDir()
	if err != nil {
		return "", err
	}

	fpath := ""
	err = filepath.WalkDir(home, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && entry.Name() == id+filepath.Ext(entry.Name()) {
			fpath = path
			return ArchiveFoundError
		}

		return nil
	})

	if err != nil && !errors.Is(err, ArchiveFoundError) {
		return "", err
	}

	return fpath, nil
}
