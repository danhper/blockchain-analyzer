package processor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/ugorji/go/codec"
)

const logInterval int = 10000

type FileFormat int

const (
	JSONFormat FileFormat = iota
	MsgpackFormat
)

var (
	msgpackHandle = &codec.MsgpackHandle{}
)

func InferFormat(filepath string) (FileFormat, error) {
	if strings.Contains(filepath, ".jsonl") {
		return JSONFormat, nil
	}
	if strings.Contains(filepath, ".dat") {
		return MsgpackFormat, nil
	}
	return JSONFormat, fmt.Errorf("invalid filename %s", filepath)
}

func YieldBlocks(reader io.Reader, blockchain core.Blockchain, format FileFormat) <-chan core.Block {
	stream := bufio.NewReader(reader)
	blocks := make(chan core.Block)

	var decoder *codec.Decoder
	if format == MsgpackFormat {
		decoder = codec.NewDecoder(stream, msgpackHandle)
	}

	go func() {
		defer close(blocks)

		for i := 0; ; i++ {
			if i%logInterval == 0 {
				log.Printf("processed: %d", i)
			}
			block := blockchain.EmptyBlock()
			var err error
			switch format {
			case JSONFormat:
				rawLine, err := stream.ReadBytes('\n')
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Printf("failed to read line %s\n", err.Error())
					return
				}
				rawLine = bytes.ToValidUTF8(rawLine, []byte{})
				block, err = blockchain.ParseBlock(rawLine)
			case MsgpackFormat:
				err = decoder.Decode(&block)
			}

			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("could not parse: %s", err.Error())
				continue
			}

			if block != nil {
				blocks <- block
			}
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
	uniqueBlocks := make(chan core.Block)

	processed := 0
	fileDone := make(chan bool)

	var wg sync.WaitGroup
	run := core.MakeFileProcessor(func(filename string) error {
		defer wg.Done()
		fileFormat, err := InferFormat(filename)
		if err != nil {
			return err
		}
		reader, err := core.OpenFile(filename)
		if err != nil {
			return err
		}
		defer reader.Close()
		for block := range YieldBlocks(reader, blockchain, fileFormat) {
			if (start == 0 || block.Number() >= start) &&
				(end == 0 || block.Number() <= end) {
				blocks <- block
			}
		}
		fileDone <- true
		return err
	})

	seen := make(map[uint64]bool)
	go func() {
		for block := range blocks {
			if _, ok := seen[block.Number()]; !ok {
				uniqueBlocks <- block
				seen[block.Number()] = true
			}
		}
		close(uniqueBlocks)
	}()

	for _, filename := range files {
		wg.Add(1)
		go run(filename)
	}

	go func() {
		for range fileDone {
			processed++
			log.Printf("files processed: %d/%d", processed, len(files))
		}
	}()

	go func() {
		wg.Wait()
		close(blocks)
		close(fileDone)
	}()

	return uniqueBlocks, nil
}

func ComputeMissingBlockNumbers(blockNumbers map[uint64]bool, start, end uint64) []uint64 {
	missing := make([]uint64, 0)
	for blockNumber := start; blockNumber <= end; blockNumber++ {
		if _, ok := blockNumbers[blockNumber]; !ok {
			missing = append(missing, blockNumber)
		}
	}

	return missing
}

func OutputAllMissingBlockNumbers(
	blockchain core.Blockchain, globPattern string,
	outputPath string, start, end uint64) error {

	blocks, err := YieldAllBlocks(globPattern, blockchain, start, 0)
	if err != nil {
		return err
	}

	outputFile, err := core.CreateFile(outputPath)
	defer outputFile.Close()

	missingBlockNumbers := core.NewMissingBlocks(start, end)
	for block := range blocks {
		missingBlockNumbers.AddBlock(block)
	}

	missing := missingBlockNumbers.Compute()
	for _, number := range missing {
		fmt.Fprintf(outputFile, "{\"block\": %d}\n", number)
	}

	return nil
}

func CountTransactions(blockchain core.Blockchain, globPattern string, start, end uint64) (int, error) {
	blocks, err := YieldAllBlocks(globPattern, blockchain, start, end)
	if err != nil {
		return 0, err
	}
	txCounter := core.NewTransactionCounter()
	for block := range blocks {
		txCounter.AddBlock(block)
	}
	return (int)(*txCounter), nil
}

func CountActionsOverTime(
	blockchain core.Blockchain,
	globPattern string,
	start, end uint64,
	duration time.Duration,
	actionProperty core.ActionProperty) (*core.TimeGroupedActions, error) {
	blocks, err := YieldAllBlocks(globPattern, blockchain, start, end)
	if err != nil {
		return nil, err
	}
	result := core.NewTimeGroupedActions(duration, actionProperty)
	for block := range blocks {
		result.AddBlock(block)
	}
	return result, nil
}

func CountTransactionsOverTime(blockchain core.Blockchain, globPattern string,
	start, end uint64, duration time.Duration,
) (*core.TimeGroupedTransactionCount, error) {
	blocks, err := YieldAllBlocks(globPattern, blockchain, start, end)
	if err != nil {
		return nil, err
	}
	result := core.NewTimeGroupedTransactionCount(duration)
	for block := range blocks {
		result.AddBlock(block)
	}
	return result, nil
}

func GroupActions(blockchain core.Blockchain, globPattern string,
	start, end uint64, by core.ActionProperty, detailed bool,
) (*core.GroupedActions, error) {
	blocks, err := YieldAllBlocks(globPattern, blockchain, start, end)
	if err != nil {
		return nil, err
	}
	groupedActions := core.NewGroupedActions(by, detailed)
	for block := range blocks {
		groupedActions.AddBlock(block)
	}
	return groupedActions, nil
}
