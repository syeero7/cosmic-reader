package main

import (
	"encoding/json"
	"errors"
	"io"
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

		return bucket.Put([]byte("library_dir"), []byte(*dirpath))
	})
}

func (d *Database) getLibraryDir() string {
	var libraryDir []byte
	d.db.View(func(tx *bbolt.Tx) error {
		libraryDir = tx.Bucket([]byte("settings")).Get([]byte("library_dir"))
		return nil
	})

	return string(libraryDir)
}

func (d *Database) addArchive(id, fpath string) (*Archive, error) {
	arch, err := convertToCBZ(id, fpath, d.addThumbnail)
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

		return bucket.Put(ulidToBytes(id), byt)
	})

	return arch, err
}

func (d *Database) getArchive(id string) (*Archive, error) {
	arch := new(Archive)
	err := d.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("archives"))
		return json.Unmarshal(bucket.Get(ulidToBytes(id)), arch)
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

			archives[ulidToString(k)] = arch
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

	return d.db.Update(func(tx *bbolt.Tx) error {
		uid := ulidToBytes(id)
		if err := tx.Bucket([]byte("archives")).Delete(uid); err != nil {
			return err
		}

		return tx.Bucket([]byte("thumbnails")).Delete(uid)
	})
}

func (d *Database) addThumbnail(id string, file io.Reader) error {
	img, err := createThumbnail(file)
	if err != nil {
		return err
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("thumbnails"))
		if err != nil {
			return err
		}

		return bucket.Put(ulidToBytes(id), img)
	})
}

func (d *Database) getThumbnail(id string) ([]byte, error) {
	byt := []byte{}
	err := d.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("thumbnails"))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		byt = bucket.Get(ulidToBytes(id))
		if byt == nil {
			return ThumbnailNotFoundError
		}

		return nil
	})

	return byt, err
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
