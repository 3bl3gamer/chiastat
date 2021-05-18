package main

import "github.com/go-pg/migrations/v8"

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		return execSome(db, `
			ALTER TABLE nodes ADD COLUMN country text;
			ALTER TABLE raw_nodes ADD COLUMN country text;

			CREATE TABLE chiastat.node_stats (
				id serial PRIMARY KEY,

				count_total int NOT NULL,
				count_total_raw int NOT NULL,

				active_count_hours jsonb NOT NULL,
				active_count_hours_raw jsonb NOT NULL,

				countries jsonb NOT NULL,
				countries_raw jsonb NOT NULL,

				ports jsonb NOT NULL,
				ports_raw jsonb NOT NULL,

				protocol_version jsonb NOT NULL,
				software_version jsonb NOT NULL,
				node_types jsonb NOT NULL,

				created_at timestamptz NOT NULL DEFAULT now()
			);
			`)
	}, func(db migrations.DB) error {
		return execSome(db, `
			ALTER TABLE raw_nodes DROP COLUMN country;
			ALTER TABLE nodes DROP COLUMN country;
			DROP TABLE chiastat.node_stats;
			`)
	})
}
