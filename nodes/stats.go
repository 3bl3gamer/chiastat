package nodes

import (
	"chiastat/utils"

	"github.com/ansel1/merry"
)

func CMDSaveStats() error {
	db := utils.MakePGConnection()

	_, err := db.Exec(`
		INSERT INTO node_stats (
			count_total,
			count_total_raw,

			active_count_hours,
			active_count_hours_raw,

			countries,
			countries_raw,

			ports,
			ports_raw,

			protocol_version,
			software_version,
			node_types
		) VALUES ((
			SELECT count(*) FROM nodes
		), (
			SELECT count(*) FROM raw_nodes
		), (
			SELECT json_object_agg(
				hours, (SELECT count(*) FROM nodes WHERE updated_at > NOW() - hours::float * INTERVAL '1 hour')
				ORDER BY hours
			)
			FROM unnest(ARRAY[12, 24, 48, 72, 96, 120]) AS hours
		), (
			SELECT json_object_agg(
				hours, (SELECT count(*) FROM raw_nodes WHERE updated_at > NOW() - hours::float * INTERVAL '1 hour')
				ORDER BY hours
			)
			FROM unnest(ARRAY[12, 24, 48, 72, 96, 120]) AS hours
		), (
			SELECT json_object_agg(country, cnt) FROM (
				SELECT COALESCE(country, '<unknown>') AS country, count(*) AS cnt
				FROM nodes
				WHERE updated_at > NOW() - INTERVAL '1 day'
				GROUP BY country
			) AS t
		), (
			SELECT json_object_agg(country, cnt) FROM (
				SELECT COALESCE(country, '<unknown>') AS country, count(*) AS cnt
				FROM raw_nodes
				WHERE updated_at > NOW() - INTERVAL '1 day'
				GROUP BY country
			) AS t
		), (
			SELECT json_object_agg(port, cnt) FROM (
				SELECT port, count(*) AS cnt
				FROM nodes
				WHERE updated_at > NOW() - INTERVAL '1 day'
				GROUP BY port
				ORDER BY cnt DESC
				LIMIT 100
			) AS t
		), (
			SELECT json_object_agg(port, cnt) FROM (
				SELECT port, count(*) AS cnt
				FROM raw_nodes
				WHERE updated_at > NOW() - INTERVAL '1 day'
				GROUP BY port
				ORDER BY cnt DESC
				LIMIT 100
			) AS t
		), (
			SELECT json_object_agg(protocol_version, cnt) FROM (
				SELECT protocol_version, count(*) AS cnt
				FROM nodes
				WHERE updated_at > NOW() - INTERVAL '1 day'
				GROUP BY protocol_version
			) AS t
		), (
			SELECT json_object_agg(software_version, cnt) FROM (
				SELECT software_version, count(*) AS cnt
				FROM nodes
				WHERE updated_at > NOW() - INTERVAL '1 day'
				GROUP BY software_version
			) AS t
		), (
			SELECT json_object_agg(node_type, cnt) FROM (
				SELECT node_type, count(*) AS cnt
				FROM nodes
				WHERE updated_at > NOW() - INTERVAL '1 day'
				GROUP BY node_type
			) AS t
		))`)
	return merry.Wrap(err)
}
