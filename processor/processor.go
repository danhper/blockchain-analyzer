package processor

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"math"

	"github.com/danhper/blockchain-data-fetcher/core"
)

func ComputeBlockNumbers(reader io.Reader, blockchain core.Blockchain) map[uint64]bool {
	stream := bufio.NewReader(reader)
	blockNumbers := make(map[uint64]bool)
	for i := 0; ; i++ {
		if i%1000 == 0 {
			log.Printf("processed: %d", i)
		}
		rawLine, err := stream.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("failed to read line %s\n", err.Error())
			continue
		}
		rawLine = bytes.ToValidUTF8(rawLine, []byte{})
		block, err := blockchain.ParseBlock(rawLine)
		if err != nil {
			log.Printf("could not parse: %s, line: %s", err.Error(), rawLine)
			continue
		}
		blockNumbers[block.Number()] = true
	}
	return blockNumbers
}

func GetMissingBlockNumbers(reader io.Reader, blockchain core.Blockchain) (missing []uint64) {
	blockNumbers := ComputeBlockNumbers(reader, blockchain)
	minNumber, maxNumber := uint64(math.MaxUint64), uint64(0)
	for blockNumber := range blockNumbers {
		if blockNumber > maxNumber {
			maxNumber = blockNumber
		}
		if blockNumber < minNumber {
			minNumber = blockNumber
		}
	}

	for blockNumber := minNumber; blockNumber <= maxNumber; blockNumber++ {
		if _, ok := blockNumbers[blockNumber]; !ok {
			missing = append(missing, blockNumber)
		}
	}

	return missing
}
