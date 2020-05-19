package processor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"path/filepath"
	"sync"

	"github.com/danhper/blockchain-data-fetcher/core"
)

func YieldBlocks(reader io.Reader, blockchain core.Blockchain) <-chan core.Block {
	stream := bufio.NewReader(reader)
	blocks := make(chan core.Block)

	go func() {
		defer close(blocks)

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
				return
			}
			rawLine = bytes.ToValidUTF8(rawLine, []byte{})
			block, err := blockchain.ParseBlock(rawLine)
			if err != nil {
				log.Printf("could not parse: %s, line: %s", err.Error(), rawLine)
				continue
			}
			blocks <- block
		}
	}()

	return blocks
}

func YieldAllBlocks(
	globPattern string,
	blockchain core.Blockchain,
	start, end uint64) (<-chan core.Block, error) {
	files, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, err
	}

	log.Printf("starting for %d files", len(files))
	blocks := make(chan core.Block)

	var wg sync.WaitGroup
	run := core.MakeFileProcessor(func(filename string) error {
		defer wg.Done()
		reader, err := core.OpenFile(filename)
		if err != nil {
			return err
		}
		defer reader.Close()
		for block := range YieldBlocks(reader, blockchain) {
			if (start == 0 || block.Number() >= start) && (end == 0 || block.Number() <= end) {
				blocks <- block
			}
		}
		return err
	})

	for _, filename := range files {
		wg.Add(1)
		go run(filename)
	}

	go func() {
		wg.Wait()
		close(blocks)
	}()

	return blocks, nil
}

func ComputeBlockNumbers(reader io.Reader, blockchain core.Blockchain) map[uint64]bool {
	blockNumbers := make(map[uint64]bool)
	for block := range YieldBlocks(reader, blockchain) {
		blockNumbers[block.Number()] = true
	}
	return blockNumbers
}

func ComputeMissingBlockNumbers(blockNumbers map[uint64]bool, blockchain core.Blockchain) []uint64 {
	minNumber, maxNumber := uint64(math.MaxUint64), uint64(0)
	for blockNumber := range blockNumbers {
		if blockNumber > maxNumber {
			maxNumber = blockNumber
		}
		if blockNumber < minNumber {
			minNumber = blockNumber
		}
	}

	missing := make([]uint64, 0)
	for blockNumber := minNumber; blockNumber <= maxNumber; blockNumber++ {
		if _, ok := blockNumbers[blockNumber]; !ok {
			missing = append(missing, blockNumber)
		}
	}

	return missing
}

func OutputAllMissingBlockNumbers(
	blockchain core.Blockchain, globPattern string,
	outputPath string, start uint64) error {

	blocks, err := YieldAllBlocks(globPattern, blockchain, start, 0)
	if err != nil {
		return err
	}

	outputFile, err := core.CreateFile(outputPath)
	defer outputFile.Close()

	blockNumbers := make(map[uint64]bool)
	for block := range blocks {
		blockNumbers[block.Number()] = true
	}

	missing := ComputeMissingBlockNumbers(blockNumbers, blockchain)
	for _, number := range missing {
		fmt.Fprintf(outputFile, "{\"block\": %d}\n", number)
	}

	return nil
}

func CountTransactions(blockchain core.Blockchain, globPattern string, start, end uint64) (int, error) {
	totalCount := 0
	blocks, err := YieldAllBlocks(globPattern, blockchain, start, end)
	if err != nil {
		return 0, err
	}
	for block := range blocks {
		totalCount += block.TransactionsCount()
	}
	return totalCount, nil
}
