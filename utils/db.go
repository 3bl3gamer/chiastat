package utils

import (
	"context"
	"database/sql"
	"os"

	"github.com/abh/geoip"
	"github.com/ansel1/merry"
	"github.com/go-pg/pg/v10"
)

func MakePGConnection() *pg.DB {
	db := pg.Connect(&pg.Options{User: "chiastat", Password: "chia", Database: "chiastat_db"})
	// db.AddQueryHook(dbLogger{})
	return db
}

func MakeGeoIPConnection() (*geoip.GeoIP, *geoip.GeoIP, error) {
	gdb, err := geoip.Open("/usr/share/GeoIP/GeoIP.dat")
	if err != nil {
		return nil, nil, merry.Wrap(err)
	}
	gdb6, err := geoip.Open("/usr/share/GeoIP/GeoIPv6.dat")
	if err != nil {
		return nil, nil, merry.Wrap(err)
	}
	return gdb, gdb6, nil
}

func SaveChunked(db *pg.DB, chunkSize int, channel chan interface{}, handler func(tx *pg.Tx, items []interface{}) error) error {
	var err error
	ctx := context.Background()
	items := make([]interface{}, 0, chunkSize)
	for item := range channel {
		items = append(items, item)
		if len(items) >= chunkSize {
			err = db.RunInTransaction(ctx, func(tx *pg.Tx) error {
				return merry.Wrap(handler(tx, items))
			})
			if err != nil {
				return merry.Wrap(err)
			}
			items = items[:0]
		}
	}
	if len(items) > 0 {
		err = db.RunInTransaction(ctx, func(tx *pg.Tx) error {
			return merry.Wrap(handler(tx, items))
		})
	}
	return merry.Wrap(err)
}

func IsPGDeadlock(err error) bool {
	if perr, ok := merry.Unwrap(err).(pg.Error); ok {
		return perr.Field('C') == "40P01"
	}
	return false
}

func OpenExistingSqlite3(dbPath string) (*sql.DB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, merry.Errorf("not found: %s", dbPath)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	return db, nil
}
