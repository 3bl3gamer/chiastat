package utils

import (
	"context"

	"github.com/ansel1/merry"
	"github.com/go-pg/pg/v10"
)

func MakePGConnection() *pg.DB {
	db := pg.Connect(&pg.Options{User: "chiastat", Password: "chia", Database: "chiastat_db"})
	// db.AddQueryHook(dbLogger{})
	return db
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
