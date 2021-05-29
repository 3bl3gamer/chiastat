package chia

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ansel1/merry"
)

type Scanner interface {
	Scan(dest ...interface{}) error
}

func BlockRecordFromRow(rows Scanner) (*BlockRecord, error) {
	var blockBytes []byte
	err := rows.Scan(&blockBytes)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	buf := NewParseBuf(blockBytes)
	br := BlockRecordFromBytes(buf)
	buf.ensureEmpty()
	if buf.err != nil {
		return nil, buf.err
	}
	return &br, nil
}

func BlockRecordByHeight(db *sql.DB, height int64) (*BlockRecord, error) {
	row := db.QueryRow("SELECT block FROM block_records WHERE height = ?", height)
	return BlockRecordFromRow(row)
}

func FullBlockFromRow(row Scanner) (*FullBlock, error) {
	var blockBytes []byte
	if err := row.Scan(&blockBytes); err != nil {
		return nil, merry.Wrap(err)
	}
	buf := NewParseBuf(blockBytes)
	block := FullBlockFromBytes(buf)
	buf.ensureEmpty()
	if buf.err != nil {
		return nil, merry.Wrap(buf.err)
	}
	return &block, nil
}

func FullBlockByHeight(db *sql.DB, height uint32) (*FullBlock, error) {
	row := db.QueryRow("SELECT block FROM full_blocks WHERE height = ?", height)
	return FullBlockFromRow(row)
}

func estimateNetworkSpaceInner(weight0, weight1 *big.Int, totalIters0, totalIters1 *big.Int) *big.Int {
	// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/rpc/full_node_rpc_api.py#L276
	deltaWeight := (&big.Int{}).Sub(weight1, weight0)
	deltaIters := (&big.Int{}).Sub(totalIters1, totalIters0)
	additionalDifficultyConstant := (&big.Int{}).Exp(big.NewInt(2), big.NewInt(67), nil)
	eligiblePlotsFilterMultiplier := (&big.Int{}).Exp(big.NewInt(2), big.NewInt(9), nil)
	UI_ACTUAL_SPACE_CONSTANT_FACTOR := 0.762

	spaceEstimate := &big.Int{}
	spaceEstimate.Mul(additionalDifficultyConstant, eligiblePlotsFilterMultiplier)
	spaceEstimate.Mul(spaceEstimate, deltaWeight)
	spaceEstimate.Div(spaceEstimate, deltaIters)
	spaceEstimate.Mul(spaceEstimate, big.NewInt(int64(UI_ACTUAL_SPACE_CONSTANT_FACTOR*1000000)))
	spaceEstimate.Div(spaceEstimate, big.NewInt(1000000))
	return spaceEstimate
}
func EstimateNetworkSpace(br0, br1 *BlockRecord) *big.Int {
	return estimateNetworkSpaceInner(br0.Weight, br1.Weight, br0.TotalIters, br1.TotalIters)
}
func EstimateNetworkSpaceFull(fb0, fb1 *FullBlock) *big.Int {
	rcb0 := &fb0.RewardChainBlock
	rcb1 := &fb1.RewardChainBlock
	return estimateNetworkSpaceInner(rcb0.Weight, rcb1.Weight, rcb0.TotalIters, rcb1.TotalIters)
}

func EstimateNetworkSpaceFromDB(db *sql.DB, lastHeight, pastOffset int64) (*big.Int, error) {
	if lastHeight < 0 {
		err := db.QueryRow("SELECT max(height) FROM block_records").Scan(&lastHeight)
		if err != nil {
			return nil, merry.Wrap(err)
		}
	}

	firstHeight := lastHeight - pastOffset
	if firstHeight < 0 {
		firstHeight = 0
	}

	br0, err := BlockRecordByHeight(db, firstHeight)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	br1, err := BlockRecordByHeight(db, lastHeight)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	// fmt.Println(br0.Timestamp, time.Unix(int64(br0.Timestamp), 0))
	// fmt.Println(br1.Timestamp, time.Unix(int64(br1.Timestamp), 0))

	// fb, err := FullBlockByHeight(db, firstHeight)
	// if err != nil {
	// 	fmt.Println(firstHeight)
	// 	return nil, merry.Wrap(err)
	// }
	// fmt.Printf("%d %d %#v\n", firstHeight, fb.RewardChainBlock.Height, fb.TransactionsGeneratorRefList)

	return EstimateNetworkSpace(br0, br1), nil
}

type RingBuf struct {
	items []interface{}
	pos   int
}

func NewRingBuf(size int) *RingBuf {
	return &RingBuf{make([]interface{}, 0, size), 0}
}

func (buf *RingBuf) Add(item interface{}) {
	if len(buf.items) < cap(buf.items) {
		buf.pos = len(buf.items)
		buf.items = buf.items[:buf.pos+1]
	} else {
		buf.pos += 1
		if buf.pos >= cap(buf.items) {
			buf.pos = 0
		}
	}
	buf.items[buf.pos] = item
}

func (buf *RingBuf) First() interface{} {
	first := buf.pos + 1
	if first >= len(buf.items) {
		first = 0
	}
	return buf.items[first]
}

func (buf *RingBuf) Last() interface{} {
	return buf.items[buf.pos]
}

type StampedRecord struct {
	Block *BlockRecord
	Stamp int64
}

func PrintNetworkSpaceChartFromDB(db *sql.DB) error {
	rows, err := db.Query("SELECT block FROM block_records ORDER BY height")
	if err != nil {
		return merry.Wrap(err)
	}
	defer rows.Close()

	blocks := NewRingBuf(4608 + 1)
	prevDay := int64(0)
	prevStampMS := int64(0)
	prevDayBlocksCount := 0
	prevDaySizePib := int64(0)
	for rows.Next() {
		br, err := BlockRecordFromRow(rows)
		if err != nil {
			return merry.Wrap(err)
		}

		// fmt.Println(len(blocks.items), cap(blocks.items), blocks.pos, blocks.Last().Height-blocks.First().Height)
		// if br.Height > 0 {
		// 	spaceEstimate := EstimateNetworkSpace(blocks.First(), blocks.Last())
		// 	pib := (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024))
		// 	fmt.Println(br.Height, pib)
		// }

		// if br.Height > 1 {
		// 	dayStartBlock := blocks.First().(StampedRecord)
		// 	dayEndBlock := blocks.Last().(StampedRecord)
		// 	spaceEstimate := EstimateNetworkSpace(dayStartBlock.Block, dayEndBlock.Block)
		// 	pib := (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024)).Int64()
		// 	daySizePibSum += pib
		// 	daySizePibCount += 1
		// }

		// if bytes.Equal(br.FarmerPuzzleHash[:], ph[:]) {
		// 	fmt.Printf("%#v\n", br.FarmerPuzzleHash)
		// 	fmt.Printf("%#v\n", br.RewardClaimsIncorporated)
		// 	for _, val := range br.RewardClaimsIncorporated {
		// 		sum += val.Amount
		// 	}
		// 	fmt.Println(sum)
		// }
		// if br.Height > 329355-10000 {
		// 	for _, coin := range br.RewardClaimsIncorporated {
		// 		// if bytes.Equal(coin.PuzzleHash[:], ph[:]) {
		// 		sum += coin.Amount
		// 		fmt.Println(sum/1000/1000/1000/1000, EncodePuzzleHash(coin.PuzzleHash, "xch"))
		// 		// if coin.Amount%1000 != 0 {
		// 		// 	fmt.Println(coin.Amount, coin.ParentCoinInfo)
		// 		// }
		// 		// }
		// 	}
		// 	// if bytes.Equal(br.FarmerPuzzleHash[:], ph[:]) {
		// 	// 	sum += 1
		// 	// 	fmt.Println(sum, sum/1000/1000/1000/1000)
		// 	// }
		// }

		stampMS := prevStampMS + 312
		if br.Timestamp != 0 {
			// fmt.Println(stampMS/1000-int64(br.Timestamp), stampMS, br.Timestamp)
			stampMS = int64(br.Timestamp) * 1000
		}
		brs := StampedRecord{br, stampMS / 1000}

		day := stampMS / (24 * 3600 * 1000)
		if day != prevDay && prevDay != 0 {
			dayStartBlock := blocks.First().(StampedRecord)
			dayEndBlock := blocks.Last().(StampedRecord)
			spaceEstimate := EstimateNetworkSpace(dayStartBlock.Block, dayEndBlock.Block)
			pib := (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024)).Int64()
			// fmt.Println(br.Height, br.Timestamp, time.Unix(int64(br.Timestamp), 0))
			// fmt.Println(prevDayBlocksCount, time.Unix(dayStartBlock.Stamp, 0), time.Unix(dayEndBlock.Stamp, 0), time.Unix(stampMS/1000, 0))
			bar := ""
			if prevDaySizePib > 0 && prevDaySizePib <= pib {
				bar = strings.Repeat("*", int((pib-prevDaySizePib)*300/prevDaySizePib))
			}
			fmt.Printf("%s\t%d\t%5.2f\t%5.1f%%\t%s\n",
				time.Unix(dayEndBlock.Stamp, 0).In(time.UTC).Format("2006-01-02 15:04:05"),
				prevDayBlocksCount,
				float64(pib)/1024,
				float64(pib-prevDaySizePib)*100/float64(prevDaySizePib),
				bar)
			prevDayBlocksCount = 0
			prevDaySizePib = pib
		}

		blocks.Add(brs)
		prevDay = day
		prevStampMS = stampMS
		prevDayBlocksCount += 1
		// blockTime := time.Unix(int64(br.Timestamp), 0)
		// day := br.Timestamp
	}
	err = rows.Err()
	if err != nil {
		return merry.Wrap(err)
	}

	return nil
}

func EvalFullBlockFromDB(db *sql.DB, height uint32) error {
	// 225698 first with transaction generator
	// 271489
	block, err := FullBlockByHeight(db, height)
	if err != nil {
		return merry.Wrap(err)
	}

	if block.TransactionsGenerator == nil {
		fmt.Printf("block %d has no transactions generator\n", height)
		return nil
	}
	fmt.Println("ref list:", block.TransactionsGeneratorRefList)

	refBlocks := make([]*FullBlock, len(block.TransactionsGeneratorRefList))
	for i, refHeight := range block.TransactionsGeneratorRefList {
		refBlocks[i], err = FullBlockByHeight(db, refHeight)
		if err != nil {
			return merry.Wrap(err)
		}
	}

	var args CLVMPair = CLVMPair{ATOM_NULL, ATOM_NULL}
	argsEnd := &args.First
	for _, block := range refBlocks {
		pair := CLVMPair{CLVMAtom{block.TransactionsGenerator.Bytes}, ATOM_NULL}
		b := [300]byte{}
		pair = CLVMPair{CLVMAtom{b[:]}, ATOM_NULL}
		*argsEnd = pair
		argsEnd = &pair.Rest
	}
	args = CLVMPair{block.TransactionsGenerator.Root, CLVMPair{args, ATOM_NULL}}
	result := RunProgram(ROM_BOOTSTRAP_GENERATOR.Root, args)

	fmt.Println("spent coins:")
	res := result.(CLVMPair).First
	for !res.Nullp() {
		cur := res.(CLVMPair).First
		spent_coin_parent_id := cur.(CLVMPair).First.(CLVMAtom).Bytes                                //bytes32
		spent_coin_puzzle_hash := cur.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom).Bytes              //bytes32
		spent_coin_amount := cur.(CLVMPair).Rest.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom).AsInt() //uint64
		// spent_coin: Coin = Coin(spent_coin_parent_id, spent_coin_puzzle_hash, spent_coin_amount)
		fmt.Println(hex.EncodeToString(spent_coin_parent_id), hex.EncodeToString(spent_coin_puzzle_hash), spent_coin_amount)
		res = res.(CLVMPair).Rest
	}

	fmt.Println("done")
	return nil
}

func ExportBlocksData(db *sql.DB, tableName string, out io.Writer) error {
	chunkSize := 10000
	lastHeight := -1
	countTotal := 0
	sizeTotal := 0
	sizeMax := 0
	blockSizeBuf := []byte{0, 0, 0, 0}

	for {
		rows, err := db.Query("SELECT height, block FROM "+tableName+" WHERE height > ? ORDER BY height LIMIT ?",
			lastHeight, chunkSize)
		if err != nil {
			return merry.Wrap(err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			var height uint64
			var blockBytes []byte
			err := rows.Scan(&height, &blockBytes)
			if err != nil {
				return merry.Wrap(err)
			}

			binary.BigEndian.PutUint32(blockSizeBuf, uint32(len(blockBytes)))
			if _, err := out.Write(blockSizeBuf); err != nil {
				return merry.Wrap(err)
			}
			if _, err := out.Write(blockBytes); err != nil {
				return merry.Wrap(err)
			}

			lastHeight = int(height)
			count += 1
			countTotal += 1
			sizeTotal += len(blockBytes)
			if len(blockBytes) > sizeMax {
				sizeMax = len(blockBytes)
			}
		}
		if err := rows.Err(); err != nil {
			return merry.Wrap(err)
		}

		if count < chunkSize {
			break
		}
		if countTotal%(chunkSize*2) == 0 {
			log.Printf("%d blocks, %.1f MiB, %.2f KiB/block (max %.1f KiB)",
				countTotal, float64(sizeTotal)/(1024*1024),
				float64(sizeTotal)/float64(countTotal)/1024, float64(sizeMax)/1024)
		}
	}

	return nil
}
