package nodes

import (
	"bufio"
	"chiastat/utils"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ansel1/merry"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/types"
)

type NodeAddr struct {
	Host string
	Port int64
}

type Node struct {
	ID              []byte
	Host            string
	Port            int64
	ProtocolVersion string
	SoftwareVersion string
	NodeType        string
}

type NodeAddrListAsPGTuple []*NodeAddr

func (l NodeAddrListAsPGTuple) AppendValue(b []byte, flags int) ([]byte, error) {
	// flags: https://github.com/go-pg/pg/blob/c9ee578a38d6866649072df18a3dbb36ff369747/types/flags.go
	for i, item := range l {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '(')
		b = types.AppendString(b, item.Host, 1) //quoteFlag=1
		b = append(b, ',')
		b = append(b, []byte(strconv.FormatInt(item.Port, 10))...)
		b = append(b, ')')
	}
	return b, nil
}

type NodeListAsPGIDs []*Node

func (l NodeListAsPGIDs) AppendValue(b []byte, flags int) ([]byte, error) {
	// flags: https://github.com/go-pg/pg/blob/c9ee578a38d6866649072df18a3dbb36ff369747/types/flags.go
	idBuf := make([]byte, 80)
	for i, item := range l {
		if i > 0 {
			b = append(b, ',')
		}
		idLen := hex.Encode(idBuf, item.ID)
		b = append(b, []byte("'\\x")...)
		b = append(b, idBuf[:idLen]...)
		b = append(b, '\'')
	}
	return b, nil
}

func startOldNodesLoader(db *pg.DB, nodesChan chan *NodeAddr, chunkSize int) utils.Worker {
	worker := utils.NewSimpleWorker(1)

	rawNodesChunkSize := chunkSize
	nodesChunkSize := chunkSize / 8
	if nodesChunkSize == 0 {
		nodesChunkSize = 1
	}

	go func() {
		defer worker.Done()
		ctx := context.Background()
		nodes := make([]*Node, nodesChunkSize)
		rawNodes := make([]*NodeAddr, rawNodesChunkSize)
		for {
			nodes := nodes[:0]
			rawNodes := rawNodes[:0]
			err := db.RunInTransaction(ctx, func(tx *pg.Tx) error {
				_, err := tx.Query(&rawNodes, `
					SELECT host, port FROM raw_nodes
					WHERE checked_at IS NULL
					   OR (checked_at < now() - INTERVAL '5 hours'
					       AND updated_at > now() - INTERVAL '7 days')
					ORDER BY checked_at ASC NULLS FIRST
					LIMIT ?
					FOR NO KEY UPDATE`, chunkSize)
				if utils.IsPGDeadlock(err) {
					return nil
				}
				if err != nil {
					return merry.Wrap(err)
				}
				if len(rawNodes) == 0 {
					return nil
				}
				_, err = tx.Exec(`
					UPDATE raw_nodes SET checked_at = NOW() WHERE (host,port) IN (?)`,
					NodeAddrListAsPGTuple(rawNodes))
				return merry.Wrap(err)
			})
			if err != nil {
				worker.AddError(err)
				return
			}
			err = db.RunInTransaction(ctx, func(tx *pg.Tx) error {
				_, err := tx.Query(&nodes, `
					SELECT id, host, port FROM nodes
					WHERE checked_at IS NULL
					   OR (checked_at < now() - INTERVAL '5 hours'
					       AND updated_at > now() - INTERVAL '7 days')
					ORDER BY checked_at ASC NULLS FIRST
					LIMIT ?
					FOR NO KEY UPDATE`, chunkSize)
				if utils.IsPGDeadlock(err) {
					return nil
				}
				if err != nil {
					return merry.Wrap(err)
				}
				if len(nodes) == 0 {
					return nil
				}
				_, err = tx.Exec(`
					UPDATE nodes SET checked_at = NOW() WHERE id IN (?)`,
					NodeListAsPGIDs(nodes))
				return merry.Wrap(err)
			})
			if err != nil {
				worker.AddError(err)
				return
			}

			log.Printf("UPDATE: nodes=%d, raw=%d", len(nodes), len(rawNodes))
			if len(nodes) == 0 && len(rawNodes) == 0 {
				time.Sleep(10 * time.Second)
			}
			for _, node := range rawNodes {
				nodesChan <- node
			}
			for _, node := range nodes {
				nodesChan <- &NodeAddr{Host: node.Host, Port: node.Port}
			}
		}
	}()
	return worker
}

func startNodesChecker(db *pg.DB, nodesInChan chan *NodeAddr, nodesOutChan chan *Node, rawNodesOutChan chan *NodeAddr) utils.Worker {
	worker := utils.NewSimpleWorker(2)

	inPacketNum := int64(-1)
	applyPacketNum := func(num string) error {
		packetNum, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			return merry.Wrap(err)
		}
		if inPacketNum == -1 && packetNum != inPacketNum+1 {
			log.Printf("UPDATE: WARN: expected packet num %d, got %d (%d)",
				inPacketNum+1, packetNum, packetNum-(inPacketNum+1))
		}
		inPacketNum = packetNum
		return nil
	}
	handleMessage := func(s string) error {
		switch s[0] {
		case 'H':
			items := strings.Split(s, " ")
			if len(items) != 8 {
				return merry.Errorf("expected 8 items, got %d: %s", len(items), s)
			}
			if err := applyPacketNum(items[1]); err != nil {
				return merry.Wrap(err)
			}
			id, err := hex.DecodeString(items[2])
			if err != nil {
				return merry.Wrap(err)
			}
			host := items[3]
			port, err := strconv.ParseInt(items[4], 10, 64)
			if err != nil {
				return merry.Wrap(err)
			}
			nodesOutChan <- &Node{
				ID:              id,
				Host:            host,
				Port:            port,
				ProtocolVersion: items[5],
				SoftwareVersion: items[6],
				NodeType:        items[7],
			}
		case 'R':
			items := strings.Split(s, " ")
			if len(items)%2 != 0 {
				return merry.Errorf("expected even items count, got %d: %s", len(items), s)
			}
			if err := applyPacketNum(items[1]); err != nil {
				return merry.Wrap(err)
			}
			tx, err := db.Begin()
			if err != nil {
				return merry.Wrap(err)
			}
			defer tx.Rollback()
			for i := 2; i < len(items); i += 2 {
				host := items[i]
				port, err := strconv.ParseInt(items[i+1], 10, 64)
				if err != nil {
					return merry.Wrap(err)
				}
				rawNodesOutChan <- &NodeAddr{
					Host: host,
					Port: port,
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

	stamp := time.Now().Unix()
	countMsgOut := int64(0)
	countMsgIn := int64(0)

	go func() {
		defer worker.Done()

		for {
			fReq, err := os.OpenFile("update_nodes_request.fifo", os.O_WRONLY, 0)
			if err != nil {
				worker.AddError(err)
				return
			}
			defer fReq.Close()

			for node := range nodesInChan {
				_, err := fmt.Fprintf(fReq, "C %s %d\n", node.Host, node.Port)
				if err != nil {
					log.Printf("checker send error: %s", err)
					break
				}
				if atomic.AddInt64(&countMsgOut, 1)%1000 == 0 {
					log.Printf("UPDATE: msg out=%d, in=%d, out rpm=%.1f",
						countMsgOut, countMsgIn, float64(countMsgOut)/float64(time.Now().Unix()-stamp)*60)
				}
			}
		}
	}()

	go func() {
		defer worker.Done()

		for {
			fRes, err := os.OpenFile("update_nodes_response.fifo", os.O_RDONLY, 0)
			if err != nil {
				worker.AddError(err)
				return
			}
			defer fRes.Close()

			scanner := bufio.NewScanner(fRes)
			for scanner.Scan() {
				if err := handleMessage(scanner.Text()); err != nil {
					log.Println(merry.Details(err))
				}
				atomic.AddInt64(&countMsgIn, 1)
			}
			if err := scanner.Err(); err != nil {
				worker.AddError(err)
				return
			}
		}
	}()
	return worker
}

func startNodesSaver(db *pg.DB, nodesChan chan *Node, chunkSize int) utils.Worker {
	worker := utils.NewSimpleWorker(1)
	nodesChanI := make(chan interface{}, 16)

	go func() {
		for node := range nodesChan {
			nodesChanI <- node
		}
		close(nodesChanI)
	}()

	go func() {
		defer worker.Done()
		err := utils.SaveChunked(db, chunkSize, nodesChanI, func(tx *pg.Tx, items []interface{}) error {
			for _, nodeI := range items {
				node := nodeI.(*Node)
				_, err := tx.Exec(`
					INSERT INTO nodes (id, host, port, protocol_version, software_version, node_type, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, now())
					ON CONFLICT (id) DO UPDATE SET
						host = EXCLUDED.host,
						port = EXCLUDED.port,
						protocol_version = EXCLUDED.protocol_version,
						software_version = EXCLUDED.software_version,
						node_type = EXCLUDED.node_type,
						updated_at = now()`,
					node.ID, node.Host, node.Port, node.ProtocolVersion, node.SoftwareVersion, node.NodeType,
				)
				if err != nil {
					return merry.Wrap(err)
				}
			}
			return nil
		})
		log.Println("UPDATE:SAVE:DONE")
		if err != nil {
			worker.AddError(err)
		}
	}()
	return worker
}

func startRawNodesSaver(db *pg.DB, nodesChan chan *NodeAddr, chunkSize int) utils.Worker {
	worker := utils.NewSimpleWorker(1)
	nodesChanI := make(chan interface{}, 16)

	go func() {
		for node := range nodesChan {
			nodesChanI <- node
		}
		close(nodesChanI)
	}()

	go func() {
		defer worker.Done()
		err := utils.SaveChunked(db, chunkSize, nodesChanI, func(tx *pg.Tx, items []interface{}) error {
			for _, nodeI := range items {
				node := nodeI.(*NodeAddr)
				_, err := tx.Exec(`
					INSERT INTO raw_nodes (host, port, updated_at) VALUES (?, ?, now())
					ON CONFLICT (host, port) DO UPDATE SET
						updated_at = now()`,
					node.Host, node.Port,
				)
				if err != nil {
					return merry.Wrap(err)
				}
			}
			return nil
		})
		log.Println("UPDATE:SAVE:RAW:DONE")
		if err != nil {
			worker.AddError(err)
		}
	}()
	return worker
}

func CMDUpdateNodes() error {
	db := utils.MakePGConnection()
	nodesInChan := make(chan *NodeAddr, 16)
	nodesOutChan := make(chan *Node, 16)
	rawNodesOutChan := make(chan *NodeAddr, 16)

	workers := []utils.Worker{
		startOldNodesLoader(db, nodesInChan, 512),
		startNodesChecker(db, nodesInChan, nodesOutChan, rawNodesOutChan),
		startNodesSaver(db, nodesOutChan, 32),
		startRawNodesSaver(db, rawNodesOutChan, 512),
	}
	for {
		for _, worker := range workers {
			if err := worker.PopError(); err != nil {
				return err
			}
		}
		time.Sleep(time.Second)
	}
}
