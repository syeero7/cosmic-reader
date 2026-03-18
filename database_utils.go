package main

import (
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

func (d *Database) compactDB() error {
	lastCompacted, err := d.getLastCompacted()
	if err != nil {
		return err
	}

	if twoWeeks := time.Hour * 24 * 7 * 2; time.Since(lastCompacted) < twoWeeks {
		return nil
	}

	opath, npath := "", ""
	err = func() error {
		tmp, err := bbolt.Open(filepath.Join(filepath.Dir(d.db.Path()), "tmp.db"), 0600, nil)
		if err != nil {
			return err
		}

		defer d.close()
		defer tmp.Close()
		opath, npath = d.db.Path(), tmp.Path()
		if err := bbolt.Compact(tmp, d.db, 64*1024); err != nil {
			return err
		}

		return d.updateLastCompacted(tmp, time.Now())
	}()
	if err != nil {
		return err
	}

	return os.Rename(opath, npath)
}

func (d *Database) getLastCompacted() (time.Time, error) {
	var lastCompacted time.Time
	err := d.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("settings"))
		if err != nil {
			return err
		}

		key := []byte("last_compacted")
		if byt := bucket.Get(key); byt != nil {
			return lastCompacted.UnmarshalBinary(byt)
		}

		lastCompacted = time.Now()
		byt, err := lastCompacted.MarshalBinary()
		if err != nil {
			return err
		}

		return bucket.Put(key, byt)
	})

	return lastCompacted, err
}

func (d *Database) updateLastCompacted(db *bbolt.DB, timestamp time.Time) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("settings"))
		byt, err := timestamp.MarshalBinary()
		if err != nil {
			return err
		}

		return bucket.Put([]byte("last_compacted"), byt)
	})
}
