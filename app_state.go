package main

import (
	"encoding/json"
	"fmt"
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

	if _, ok := state.Archives[id]; !ok {
		return fmt.Errorf("archive: %s not found", id)
	}

	delete(state.Archives, id)
	return s.save(state)
}

func (s *StateManager) addArchive(id string, archive Archive, replace bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.getState()
	if err != nil {
		return err
	}

	_, ok := state.Archives[id]
	if !replace && ok {
		return fmt.Errorf("archive: %s exists", id)
	}

	if replace && !ok {
		return fmt.Errorf("archive: %s not found", id)
	}

	state.Archives[id] = archive
	return s.save(state)
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
