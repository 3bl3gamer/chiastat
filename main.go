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
	"time"

	"github.com/ansel1/merry"
	_ "github.com/mattn/go-sqlite3"
)

func CMDListenPeers() error {
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
					INSERT INTO raw_nodes (host, port, updated_at) VALUES (?, ?, now())
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

func CMDImportPeers() error {
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

func CMDEstimateSize() error {
	dbPath := flag.String("db-path", utils.HomeDirOrEmpty("/.chia/mainnet/db/")+"blockchain_v1_mainnet.sqlite", "path to blockchain_v1_mainnet.sqlite")
	flag.Parse()

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		return merry.Errorf("not found: %s", *dbPath)
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	spaceEstimate, err := EstimateNetworkSpace(db, 280000, 4608)
	if err != nil {
		return merry.Wrap(err)
	}
	fmt.Printf("%dB %dPiB\n", spaceEstimate, (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024)))
	return nil
}

var commands = map[string]func() error{
	"listen-peers":  CMDListenPeers,
	"import-peers":  CMDImportPeers,
	"estimate-size": CMDEstimateSize,
}

func printUsage() {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	fmt.Printf("usage: %s [%s]", os.Args[0], strings.Join(names, "|"))
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(2)
	}
	cmd, ok := commands[os.Args[1]]
	if !ok {
		printUsage()
		os.Exit(1)
	}

	os.Args = os.Args[1:]
	if err := cmd(); err != nil {
		println(merry.Details(err))
		os.Exit(1)
	}
}
