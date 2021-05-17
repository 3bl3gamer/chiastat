package nodes

import (
	"bufio"
	"chiastat/utils"
	"encoding/hex"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/ansel1/merry"
)

func CMDListenNodes() error {
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
					INSERT INTO nodes (id, host, port, protocol_version, software_version, node_type, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, now())
					ON CONFLICT (id) DO UPDATE SET
						host = EXCLUDED.host,
						port = EXCLUDED.port,
						protocol_version = EXCLUDED.protocol_version,
						software_version = EXCLUDED.software_version,
						node_type = EXCLUDED.node_type,
						updated_at = now()`,
				id, host, port, items[4], items[5], items[6],
			)
			if err != nil {
				return merry.Wrap(err)
			}
		// case 'R':
		// 	tx, err := db.Begin()
		// 	if err != nil {
		// 		return merry.Wrap(err)
		// 	}
		// 	defer tx.Rollback()
		// 	items := strings.Split(s, " ")
		// 	if len(items)%2 != 1 {
		// 		return merry.Errorf("expected odd items count, got %d: %s", len(items), s)
		// 	}
		// 	for i := 1; i < len(items); i += 2 {
		// 		host := items[i]
		// 		port, err := strconv.ParseInt(items[i+1], 10, 64)
		// 		if err != nil {
		// 			return merry.Wrap(err)
		// 		}
		// 		_, err = tx.Exec(`
		// 			INSERT INTO raw_nodes (host, port, updated_at) VALUES (?, ?, now())
		// 			ON CONFLICT (host, port) DO UPDATE SET
		// 				updated_at = now()`,
		// 			host, port,
		// 		)
		// 		if err != nil {
		// 			return merry.Wrap(err)
		// 		}
		// 	}
		// 	if err := tx.Commit(); err != nil {
		// 		return merry.Wrap(err)
		// 	}
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
