package main

import "github.com/go-pg/migrations/v8"

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		return execSome(db, `
			DROP INDEX nodes__created_at__nulls_fisrt__index;
			DROP INDEX raw_nodes__created_at__nulls_fisrt__index;

			ALTER TABLE nodes ADD COLUMN seems_off bool NOT NULL DEFAULT false;
			ALTER TABLE raw_nodes ADD COLUMN seems_off bool NOT NULL DEFAULT false;

			UPDATE nodes SET seems_off = true WHERE checked_at IS NOT NULL AND updated_at < NOW() - INTERVAL '7 days';
			UPDATE raw_nodes SET seems_off = true WHERE checked_at IS NOT NULL AND updated_at < NOW() - INTERVAL '7 days';

			CREATE INDEX nodes__seems_off_and_checked_at__index ON nodes (checked_at nulls first) WHERE NOT seems_off;
			CREATE INDEX raw_nodes__seems_off_and_checked_at__index ON raw_nodes (checked_at nulls first) WHERE NOT seems_off;
			`)
	}, func(db migrations.DB) error {
		return execSome(db, `
			CREATE INDEX nodes__created_at__nulls_fisrt__index ON nodes (checked_at nulls first);
			CREATE INDEX raw_nodes__created_at__nulls_fisrt__index ON raw_nodes (checked_at nulls first);

			ALTER TABLE nodes DROP COLUMN seems_off;
			ALTER TABLE raw_nodes DROP COLUMN seems_off;
			`)
	})
}
