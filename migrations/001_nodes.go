package main

import "github.com/go-pg/migrations/v8"

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		return execSome(db, `
			CREATE TABLE chiastat.nodes (
				id bytea PRIMARY KEY,
				host text NOT NULL,
				port int NOT NULL,
				protocol_version text NOT NULL,
				software_version text NOT NULL,
				node_type text NOT NULL,
				created_at timestamptz NOT NULL DEFAULT now(),
				checked_at timestamptz,
				updated_at timestamptz,
				CHECK (length(id) = 32)
			);
			CREATE TABLE chiastat.raw_nodes (
				host text NOT NULL,
				port int NOT NULL,
				PRIMARY KEY (host, port)
			);
			`)
	}, func(db migrations.DB) error {
		return execSome(db, `
			DROP TABLE chiastat.nodes;
			DROP TABLE chiastat.raw_nodes;
			`)
	})
}
