package nodes

import (
	"chiastat/chia/network"
	"chiastat/chia/types"
	chiautils "chiastat/chia/utils"
	"chiastat/utils"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/abh/geoip"
	"github.com/ansel1/merry"
	"github.com/go-pg/pg/v10"
	pgtypes "github.com/go-pg/pg/v10/types"
)

func JoinHostPort(host string, port uint16) string {
	return net.JoinHostPort(host, strconv.Itoa(int(port)))
}

type NodeAddr struct {
	Host    string
	Port    uint16
	Country *string
}

type Node struct {
	ID              []byte
	Host            string
	Port            uint16
	ProtocolVersion string
	SoftwareVersion string
	NodeType        string
	Country         *string
}

type NodeAddrListAsPGTuple []*NodeAddr

func (l NodeAddrListAsPGTuple) AppendValue(b []byte, flags int) ([]byte, error) {
	// flags: https://github.com/go-pg/pg/blob/c9ee578a38d6866649072df18a3dbb36ff369747/types/flags.go
	for i, item := range l {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '(')
		b = pgtypes.AppendString(b, item.Host, 1) //quoteFlag=1
		b = append(b, ',')
		b = append(b, []byte(strconv.FormatInt(int64(item.Port), 10))...)
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

func startNodesChecker(db *pg.DB, sslDir string, nodesInChan chan *NodeAddr, nodesOutChan chan *Node, rawNodesOutChan chan []types.TimestampedPeerInfo, concurrency int) utils.Worker {
	worker := utils.NewSimpleWorker(concurrency)

	var totalCount int64 = 0
	var totalCountOk int64 = 0
	var stampCount int64 = 0
	stamp := time.Now().Unix()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer worker.Done()

			cfg, err := network.MakeTSLConfigFromFiles(
				sslDir+"/ca/chia_ca.crt",
				sslDir+"/full_node/public_full_node.crt",
				sslDir+"/full_node/public_full_node.key")
			if err != nil {
				worker.AddError(err)
				return
			}

			handleNode := func(node *NodeAddr) error {
				c, err := network.ConnectTo(cfg, JoinHostPort(node.Host, node.Port))
				if err != nil {
					return merry.Wrap(err)
				}
				c.SetDebug(false)
				hs, err := c.PerformHandshake()
				if err != nil {
					return merry.Wrap(err)
				}
				c.StartRoutines()

				id := c.PeerID()
				nodeType, _ := types.NodeTypeName(hs.NodeType)
				for len(nodesOutChan) == cap(nodesOutChan) {
					time.Sleep(time.Second)
				}
				nodesOutChan <- &Node{
					ID:              id[:],
					Host:            c.RemoteAddr().(*net.TCPAddr).IP.String(),
					Port:            hs.ServerPort,
					ProtocolVersion: hs.ProtocolVersion,
					SoftwareVersion: hs.SoftwareVersion,
					NodeType:        nodeType,
				}

				for i := 0; i < 3; i++ {
					peers, err := c.RequestPeers()
					if err != nil {
						break
					}
					for len(rawNodesOutChan) == cap(rawNodesOutChan) {
						time.Sleep(time.Second)
					}
					rawNodesOutChan <- peers.PeerList
				}
				return nil
			}

			for node := range nodesInChan {
				err := handleNode(node)
				if err == nil {
					atomic.AddInt64(&totalCountOk, 1)
				}
				atomic.AddInt64(&stampCount, 1)
				if atomic.AddInt64(&totalCount, 1)%2500 == 0 {
					log.Printf("CHECK: nodes checked: %d, ok: %d, rpm: %.2f",
						totalCount, totalCountOk, float64(stampCount*60)/float64(time.Now().Unix()-stamp))
					stampCount = 0
					stamp = time.Now().Unix()
				}
			}
		}()
	}

	return worker
}

func startRawNodesFilter(nodeChunksChan chan []types.TimestampedPeerInfo, nodesChan chan *NodeAddr) utils.Worker {
	worker := utils.NewSimpleWorker(2)

	cleanupInterval := int64(10 * 60)
	updateInterval := int64(60 * 60)

	go func() {
		defer worker.Done()

		nodeStamps := make(map[string]int64)
		lastCleanupStamp := time.Now().Unix()
		chunksCount := 0

		for chunk := range nodeChunksChan {
			now := time.Now().Unix()
			if now-lastCleanupStamp > cleanupInterval {
				count := 0
				for addr, stamp := range nodeStamps {
					if now-stamp > updateInterval {
						delete(nodeStamps, addr)
						count += 1
					}
				}
				log.Printf("FILTER: raw nodes cleanup: %d removed, %d remaining", count, len(nodeStamps))
				lastCleanupStamp = now
			}

			for _, node := range chunk {
				addr := JoinHostPort(node.Host, node.Port)
				if stamp, ok := nodeStamps[addr]; !ok || now-stamp > updateInterval {
					nodesChan <- &NodeAddr{Host: node.Host, Port: node.Port}
					nodeStamps[addr] = now
				}
			}

			chunksCount += 1
			if chunksCount%100 == 0 {
				log.Printf("FILTER: raw nodes in filter: %d", len(nodeStamps))
			}
		}
	}()

	return worker
}

func tryGetCountry(gdb, gdb6 *geoip.GeoIP, host string, tryResolve bool) *string {
	hostIP := net.ParseIP(host)
	if hostIP == nil {
		if tryResolve {
			addrs, _ := net.LookupHost(host)
			ipFound := false
			for _, addr := range addrs {
				hostIP = net.ParseIP(addr)
				if hostIP != nil {
					host = addr
					ipFound = true
					break
				}
			}
			if !ipFound {
				return nil
			}
		} else {
			return nil
		}
	}
	if hostIP.To4() == nil {
		if code, _ := gdb6.GetCountry_v6(host); code != "" {
			return &code
		}
	} else {
		if code, _ := gdb.GetCountry(host); code != "" {
			return &code
		}
	}
	return nil
}

func startNodesLocationChecker(gdb, gdb6 *geoip.GeoIP, nodesIn, nodesOut chan *Node, rawNodesIn, rawNodesOut chan *NodeAddr, numWorkers int) utils.Worker {
	worker := utils.NewSimpleWorker(2 * numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer worker.Done()
			for node := range nodesIn {
				node.Country = tryGetCountry(gdb, gdb6, node.Host, true)
				nodesOut <- node
			}
			close(nodesOut)
		}()
	}

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer worker.Done()
			for node := range rawNodesIn {
				node.Country = tryGetCountry(gdb, gdb6, node.Host, false)
				rawNodesOut <- node
			}
			close(rawNodesOut)
		}()
	}

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
					INSERT INTO nodes (id, host, port, protocol_version, software_version, node_type, country, updated_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, now())
					ON CONFLICT (id) DO UPDATE SET
						host = EXCLUDED.host,
						port = EXCLUDED.port,
						protocol_version = EXCLUDED.protocol_version,
						software_version = EXCLUDED.software_version,
						node_type = EXCLUDED.node_type,
						country = EXCLUDED.country,
						updated_at = now()`,
					node.ID, node.Host, node.Port, node.ProtocolVersion, node.SoftwareVersion, node.NodeType, node.Country,
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
					INSERT INTO raw_nodes (host, port, country, updated_at) VALUES (?, ?, ?, now())
					ON CONFLICT (host, port) DO UPDATE SET
						country = EXCLUDED.country,
						updated_at = now()`,
					node.Host, node.Port, node.Country,
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

func startNodesListener(sslDir string, nodesChan chan *Node, rawNodesChan chan []types.TimestampedPeerInfo) utils.Worker {
	worker := utils.NewSimpleWorker(1)

	go func() {
		defer worker.Done()

		connHandler := func(c *network.WSChiaConnection) {
			fmt.Println("new connection from:", c.PeerIDHex())

			c.SetMessageHandler(func(msgID uint16, msg chiautils.FromBytes) {
				switch msg := msg.(type) {
				case *types.RequestPeers:
					c.SendReply(msgID, types.RespondPeers{PeerList: nil})
				case *types.RespondPeers:
					log.Printf("LISTEN: some peers: %d", len(msg.PeerList))
					rawNodesChan <- msg.PeerList
				case *types.NewPeak,
					*types.NewCompactVDF,
					*types.NewSignagePointOrEndOfSubSlot,
					*types.NewUnfinishedBlock,
					*types.RequestMempoolTransactions,
					*types.NewTransaction:
					// do nothing
				default:
					log.Printf("LISTEN: unexpected message: %#v", msg)
				}
			})

			hs, err := c.PerformHandshake()
			if err != nil {
				log.Printf("LISTEN: handshake error: %s", err)
				c.Close()
				return
			}
			c.StartRoutines()

			id := c.PeerID()
			nodeType, _ := types.NodeTypeName(hs.NodeType)
			nodesChan <- &Node{
				ID:              id[:],
				Host:            c.RemoteAddr().(*net.TCPAddr).IP.String(),
				Port:            hs.ServerPort,
				ProtocolVersion: hs.ProtocolVersion,
				SoftwareVersion: hs.SoftwareVersion,
				NodeType:        nodeType,
			}

			for i := 0; i < 3; i++ {
				peers, err := c.RequestPeers()
				if err != nil {
					log.Printf("LISTEN: peers error: %s", err)
					break
				}
				log.Println("LISTEN: peers from incoming node:", len(peers.PeerList))
				rawNodesChan <- peers.PeerList
			}
		}

		err := network.ListenOn("0.0.0.0", sslDir+"/ca/chia_ca.crt", sslDir+"/ca/chia_ca.key", connHandler)
		if err != nil {
			worker.AddError(err)
		}
	}()
	return worker
}

func CMDUpdateNodes() error {
	sslDir := flag.String("ssl-dir", utils.HomeDirOrEmpty("/.chia/mainnet/ssl"), "path to chia/mainnet/ssl directory")
	flag.Parse()

	db := utils.MakePGConnection()
	gdb, gdb6, err := utils.MakeGeoIPConnection()
	if err != nil {
		return merry.Wrap(err)
	}

	dbNodeAddrs := make(chan *NodeAddr, 16)

	rawNodeChunks := make(chan []types.TimestampedPeerInfo, 16)

	nodesNoLoc := make(chan *Node, 16)
	rawNodesNoLoc := make(chan *NodeAddr, 16)

	nodesOut := make(chan *Node, 32)
	rawNodesOut := make(chan *NodeAddr, 256)

	workers := []utils.Worker{
		// input
		startOldNodesLoader(db, dbNodeAddrs, 512),
		startNodesChecker(db, *sslDir, dbNodeAddrs, nodesNoLoc, rawNodeChunks, 256),
		startNodesListener(*sslDir, nodesNoLoc, rawNodeChunks),
		// process
		startRawNodesFilter(rawNodeChunks, rawNodesNoLoc),
		startNodesLocationChecker(gdb, gdb6, nodesNoLoc, nodesOut, rawNodesNoLoc, rawNodesOut, 32),
		// save
		startNodesSaver(db, nodesOut, 32),
		startRawNodesSaver(db, rawNodesOut, 512),
	}
	go func() {
		for {
			log.Printf("UPDATE: chans: (%d) -> (%d, %d -> %d) -> (%d, %d)",
				len(dbNodeAddrs),
				len(nodesNoLoc), len(rawNodeChunks), len(rawNodesNoLoc),
				len(nodesOut), len(rawNodesOut))
			time.Sleep(10 * time.Second)
		}
	}()
	for {
		for _, worker := range workers {
			if err := worker.PopError(); err != nil {
				return err
			}
		}
		time.Sleep(time.Second)
	}
}
