package main

import (
	"bufio"
	"chiastat/utils"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/ansel1/merry"
	_ "github.com/mattn/go-sqlite3"
)

func mainErr() error {
	listenPeers := flag.Bool("listen-peers", false, "listen TCP peer messages from hook on port 18444")
	addPeersFrom := flag.String("add-peers-from", "", "path to peer_table_node.sqlite to load peers from")
	flag.Parse()

	if *listenPeers {
		db := utils.MakePGConnection()

		handleMessage := func(s string) error {
			switch s[0] {
			case 'H':
				items := strings.Split(s, " ")
				if len(items) != 7 {
					return merry.Errorf("expected 7 items, got %d: %s", len(items), s)
				}
				id, err := hex.DecodeString(items[1])
				if err != nil {
					return merry.Wrap(err)
				}
				host := items[2]
				port, err := strconv.ParseInt(items[3], 10, 64)
				if err != nil {
					return merry.Wrap(err)
				}
				_, err = db.Exec(`
					INSERT INTO nodes (id, host, port, protocol_version, software_version, node_type)
					VALUES (?, ?, ?, ?, ?, ?)
					ON CONFLICT (id) DO UPDATE SET
						host = EXCLUDED.host,
						port = EXCLUDED.port,
						protocol_version = EXCLUDED.protocol_version,
						software_version = EXCLUDED.software_version,
						node_type = EXCLUDED.node_type`,
					id, host, port, items[4], items[5], items[6],
				)
				if err != nil {
					return merry.Wrap(err)
				}
			case 'R':
				tx, err := db.Begin()
				if err != nil {
					return merry.Wrap(err)
				}
				defer tx.Rollback()
				items := strings.Split(s, " ")
				if len(items)%2 != 1 {
					return merry.Errorf("expected odd items count, got %d: %s", len(items), s)
				}
				for i := 1; i < len(items); i += 2 {
					host := items[i]
					port, err := strconv.ParseInt(items[i+1], 10, 64)
					if err != nil {
						return merry.Wrap(err)
					}
					_, err = tx.Exec(`
					INSERT INTO raw_nodes (host, port) VALUES (?, ?)
					ON CONFLICT (host, port) DO UPDATE SET
						updated_at = now()`,
						host, port,
					)
					if err != nil {
						return merry.Wrap(err)
					}
				}
				if err := tx.Commit(); err != nil {
					return merry.Wrap(err)
				}
			default:
				return merry.Errorf("unexpected message: %s", s)
			}
			return nil
		}

		ln, err := net.Listen("tcp", "127.0.0.1:18444")
		if err != nil {
			return merry.Wrap(err)
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				return merry.Wrap(err)
			}
			buf := bufio.NewReader(conn)
			go func() {
				for {
					msg, err := buf.ReadString('\n')
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Println(merry.Details(err))
						break
					}
					if strings.HasSuffix(msg, "\n") {
						msg = msg[:len(msg)-1]
					}

					if err := handleMessage(msg); err != nil {
						log.Println(merry.Details(err))
					}
				}
			}()
		}
	}

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
