package main

import "github.com/go-pg/migrations/v8"

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		return execSome(db, `
			CREATE INDEX nodes__created_at__nulls_fisrt__index ON nodes (checked_at nulls first);
			CREATE INDEX raw_nodes__created_at__nulls_fisrt__index ON raw_nodes (checked_at nulls first);
			`)
	}, func(db migrations.DB) error {
		return execSome(db, `
			DROP INDEX nodes__created_at__nulls_fisrt__index;
			DROP INDEX raw_nodes__created_at__nulls_fisrt__index;
			`)
	})
}
