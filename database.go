package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

type Database struct {
	db *bbolt.DB
}

type Archive struct {
	PageCount int    `json:"pageCount"`
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
}

type Settings struct {
	LibraryDir string `json:"libraryDir"`
}

func initDB() (*Database, error) {
	config, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	db, err := bbolt.Open(filepath.Join(config, "main.db"), 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	d := &Database{db: db}
	d.setLibraryDir(nil)
	return d, nil
}

func (d *Database) close() error {
	return d.db.Close()
}

func (d *Database) setLibraryDir(dirpath *string) error {
	if dirpath == nil {
		home, err := getHomeDir()
		if err != nil {
			return err
		}

		dirpath = &home
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("settings"))
		if err != nil {
			return err
		}

		settings := Settings{LibraryDir: *dirpath}
		byt, err := json.Marshal(settings)
		if err != nil {
			return err
		}

		return bucket.Put([]byte("libraryDir"), byt)
	})
}

func (d *Database) getLibraryDir() (string, error) {
	settings := new(Settings)
	err := d.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("settings"))
		return json.Unmarshal(bucket.Get([]byte("libraryDir")), settings)
	})

	return settings.LibraryDir, err
}

func (d *Database) addArchive(id, fpath string) (*Archive, error) {
	arch, err := extractComic(id, fpath)
	if err != nil {
		return nil, err
	}

	err = d.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("archives"))
		if err != nil {
			return err
		}

		byt, err := json.Marshal(arch)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(id), byt)
	})

	return arch, err
}

func (d *Database) getArchive(id string) (*Archive, error) {
	arch := new(Archive)
	err := d.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("archives"))
		return json.Unmarshal(bucket.Get([]byte(id)), arch)
	})

	return arch, err
}

func (d *Database) getAllArchives() (map[string]*Archive, error) {
	archives := make(map[string]*Archive)
	err := d.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("archives"))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			arch := new(Archive)
			if err := json.Unmarshal(v, arch); err != nil {
				return err
			}

			archives[string(k)] = arch
			return nil
		})
	})

	return archives, err
}

func (d *Database) removeArchive(id string) error {
	fpath, err := findArchive(id)
	if err != nil {
		return err
	}

	if err := os.Remove(fpath); err != nil && !os.IsNotExist(err) {
		return err
	}

	arch, err := d.getArchive(id)
	if err != nil {
		return err
	}

	if err := deleteCachedThumbnails(&arch.Thumbnail); err != nil && !os.IsNotExist(err) {
		return err
	}

	err = d.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("archives")).Delete([]byte(id))
	})

	return err
}

var ArchiveFoundError = errors.New("archive found")

// find archive path
func findArchive(id string) (string, error) {
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
