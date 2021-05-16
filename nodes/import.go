package nodes

import (
	"chiastat/utils"
	"database/sql"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ansel1/merry"
)

func CMDImportNodes() error {
	dbPath := flag.String("db-path", utils.HomeDirOrEmpty("/.chia/mainnet/db/")+"peer_table_node.sqlite", "path to peer_table_node.sqlite to load peers from")
	flag.Parse()

	count := 0

	db := utils.MakePGConnection()
	tx, err := db.Begin()
	if err != nil {
		return merry.Wrap(err)
	}

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		return merry.Errorf("not found: %s", *dbPath)
	}
	peersDB, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		return merry.Wrap(err)
	}
	defer peersDB.Close()

	insertRawNode := func(host string, port int64, createdAt time.Time) (int, error) {
		res, err := tx.Exec(`
			INSERT INTO raw_nodes (host, port, created_at) VALUES (?, ?, ?)
			ON CONFLICT (host, port) DO UPDATE SET
				created_at = least(raw_nodes.created_at, EXCLUDED.created_at)`,
			host, port, createdAt)
		return res.RowsAffected(), merry.Wrap(err)
	}

	rows, err := peersDB.Query("SELECT value FROM peer_nodes")
	if err != nil {
		return merry.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return merry.Wrap(err)
		}
		items := strings.Split(value, " ")
		if len(items) != 5 {
			log.Printf("WARN: expected 5 items, got %d: %s", len(items), value)
		}
		port1, err := strconv.ParseInt(items[1], 10, 64)
		if err != nil {
			return merry.Wrap(err)
		}
		stamp, err := strconv.ParseInt(items[2], 10, 64)
		if err != nil {
			return merry.Wrap(err)
		}
		createdAt := time.Unix(stamp, 0)
		port4, err := strconv.ParseInt(items[4], 10, 64)
		if err != nil {
			return merry.Wrap(err)
		}
		n0, err := insertRawNode(items[0], port1, createdAt)
		if err != nil {
			return merry.Wrap(err)
		}
		n1, err := insertRawNode(items[3], port4, createdAt)
		if err != nil {
			return merry.Wrap(err)
		}
		count += n0 + n1
		if count%1000 == 0 {
			log.Printf("%d nodes", count)
		}
	}
	if err = rows.Err(); err != nil {
		return merry.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return merry.Wrap(err)
	}

	log.Printf("done, %d node(s)", count)
	return nil
}
