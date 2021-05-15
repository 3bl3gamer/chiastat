package utils

import "github.com/go-pg/pg/v10"

func MakePGConnection() *pg.DB {
	db := pg.Connect(&pg.Options{User: "chiastat", Password: "chia", Database: "chiastat_db"})
	// db.AddQueryHook(dbLogger{})
	return db
}
