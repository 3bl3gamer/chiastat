package main

import (
	"chiastat/utils"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ansel1/merry"
	_ "github.com/mattn/go-sqlite3"
)

func mainErr() error {
	addPeersFrom := flag.String("add-peers-from", "", "path to peer_table_node.sqlite to load peers from")
	flag.Parse()

	if *addPeersFrom != "" {
		count := 0

		db := utils.MakePGConnection()
		tx, err := db.Begin()
		if err != nil {
			return merry.Wrap(err)
		}

		peersDB, err := sql.Open("sqlite3", *addPeersFrom)
		if err != nil {
			return merry.Wrap(err)
		}
		defer peersDB.Close()

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
			port4, err := strconv.ParseInt(items[4], 10, 64)
			if err != nil {
				return merry.Wrap(err)
			}
			res, err := tx.Exec(
				"INSERT INTO raw_nodes (host, port) VALUES (?, ?), (?, ?) ON CONFLICT DO NOTHING",
				items[0], port1, items[3], port4)
			if err != nil {
				return merry.Wrap(err)
			}
			count += res.RowsAffected()
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

	db, err := sql.Open("sqlite3", "blockchain_v1_mainnet.sqlite")
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT header_hash, prev_hash, height, block, sub_epoch_summary, is_peak, is_block FROM block_records LIMIT 10")
	if err != nil {
		return merry.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		br, err := BlockRecordFromRow(rows)
		if err != nil {
			return merry.Wrap(err)
		}
		fmt.Println(br.HeaderHash, br.Height, br.Weight, br.TotalIters)
	}
	err = rows.Err()
	if err != nil {
		return merry.Wrap(err)
	}

	spaceEstimate, err := EstimateNetworkSpace(db, 280000, 4608)
	fmt.Println(spaceEstimate, (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024)))

	return nil
}

func main() {
	if err := mainErr(); err != nil {
		println(merry.Details(err))
		os.Exit(1)
	}
}
