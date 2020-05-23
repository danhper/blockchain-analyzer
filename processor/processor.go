package processor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"path"
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
			var block core.Block
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
				blockchainBlock := blockchain.EmptyBlock()
				err = decoder.Decode(&blockchainBlock)
				block = blockchainBlock
			}

			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("could not parse: %s", err.Error())
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

	blockNumbers := make(map[uint64]bool)
	for block := range blocks {
		blockNumbers[block.Number()] = true
	}

	missing := ComputeMissingBlockNumbers(blockNumbers, start, end)
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

func CountActions(blockchain core.Blockchain, globPattern string, start, end uint64) (*core.ActionsCount, error) {
	blocks, err := YieldAllBlocks(globPattern, blockchain, start, end)
	if err != nil {
		return nil, err
	}
	actionsCount := core.NewActionsCount()
	for block := range blocks {
		actionsCount.Merge(block.GetActionsCount(core.ActionName))
	}
	return actionsCount, nil
}

func CountActionsPerTime(
	blockchain core.Blockchain,
	globPattern string,
	start, end uint64,
	duration time.Duration,
	actionProperty core.ActionProperty) (*core.GroupedActions, error) {
	blocks, err := YieldAllBlocks(globPattern, blockchain, start, end)
	if err != nil {
		return nil, err
	}
	result := core.NewGroupedActions(duration)
	for block := range blocks {
		result.AddActions(block.Time(), block.GetActionsCount(actionProperty))
	}
	return result, nil
}

func ExportToMsgpack(
	blockchain core.Blockchain,
	globPattern string,
	start, end uint64,
	outputDir string,
) error {
	files, err := filepath.Glob(globPattern)
	if err != nil {
		return err
	}
	processed := 0
	fileDone := make(chan bool)
	var wg sync.WaitGroup

	exportFile := core.MakeFileProcessor(func(filename string) error {
		defer wg.Done()
		reader, err := core.OpenFile(filename)
		if err != nil {
			return err
		}
		defer reader.Close()

		outputFilename := strings.Replace(path.Base(filename), "jsonl", "dat", 1)
		outputFilepath := path.Join(outputDir, outputFilename)
		writer, err := core.CreateFile(outputFilepath)
		if err != nil {
			return err
		}
		defer writer.Close()

		for block := range YieldBlocks(reader, blockchain, JSONFormat) {
			if (start == 0 || block.Number() >= start) &&
				(end == 0 || block.Number() <= end) {
				var rawBlock []byte
				enc := codec.NewEncoderBytes(&rawBlock, msgpackHandle)
				if err := enc.Encode(block); err != nil {
					return err
				}
				writer.Write(rawBlock)
			}
		}
		fileDone <- true
		return err
	})

	go func() {
		for range fileDone {
			processed++
			log.Printf("files processed: %d/%d", processed, len(files))
		}
	}()

	log.Printf("exporting %d files", len(files))

	for _, filename := range files {
		wg.Add(1)
		go exportFile(filename)
	}

	wg.Wait()
	close(fileDone)

	return nil
}
