package chia

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ansel1/merry"
)

type Scanner interface {
	Scan(dest ...interface{}) error
}

const BlockRecordSelectCols = "header_hash, prev_hash, height, block, sub_epoch_summary, is_peak, is_block"

func BlockRecordFromRow(rows Scanner) (*BlockRecord, error) {
	var headerHash string
	var prevHash string
	var height int64
	var block []byte
	var subEpochSummary []byte
	var isPeak int8
	var isBlock int8
	err := rows.Scan(&headerHash, &prevHash, &height, &block, &subEpochSummary, &isPeak, &isBlock)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	// fmt.Println(headerHash, prevHash, height, isPeak, isBlock)
	// fmt.Println(block)
	buf := NewParseBuf(block)
	br := BlockRecordFromBytes(buf)
	if buf.err != nil {
		return nil, buf.err
	}
	// fmt.Printf("%#v\n", br)
	// fmt.Println(br.height, br.weight, br.totalIters)
	return &br, nil
}

func BlockRecordByHeight(db *sql.DB, height int64) (*BlockRecord, error) {
	row := db.QueryRow("SELECT "+BlockRecordSelectCols+" FROM block_records WHERE height = ?", height)
	return BlockRecordFromRow(row)
}

func EstimateNetworkSpace(br0, br1 *BlockRecord) *big.Int {
	// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/rpc/full_node_rpc_api.py#L276
	deltaWeight := (&big.Int{}).Sub(br1.Weight, br0.Weight)
	deltaIters := (&big.Int{}).Sub(br1.TotalIters, br0.TotalIters)
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
	rows, err := db.Query("SELECT " + BlockRecordSelectCols + " FROM block_records ORDER BY height")
	if err != nil {
		return merry.Wrap(err)
	}
	defer rows.Close()

	// _, ph, err := DecodePuzzleHash("xch1f0ryxk6qn096hefcwrdwpuph2hm24w69jnzezhkfswk0z2jar7aq5zzpfj")
	// if err != nil {
	// 	return merry.Wrap(err)
	// }

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
