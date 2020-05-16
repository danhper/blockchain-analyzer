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

func ComputeBlockNumbers(reader io.Reader, blockchain core.Blockchain) (map[uint64]bool, error) {
	stream := bufio.NewReader(reader)
	blockNumbers := make(map[uint64]bool)
	var lastError error = nil
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
			return nil, err
		}
		rawLine = bytes.ToValidUTF8(rawLine, []byte{})
		block, err := blockchain.ParseBlock(rawLine)
		if err != nil {
			log.Printf("could not parse: %s, line: %s", err.Error(), rawLine)
			lastError = err
			continue
		}
		blockNumbers[block.Number()] = true
	}
	return blockNumbers, lastError
}

func GetMissingBlockNumbers(reader io.Reader, blockchain core.Blockchain) ([]uint64, error) {
	blockNumbers, err := ComputeBlockNumbers(reader, blockchain)
	if blockNumbers == nil {
		return []uint64{}, err
	}
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

	return missing, err
}

func OutputAllMissingBlockNumbers(blockchain core.Blockchain, globPattern string, outputPath string) error {
	files, err := filepath.Glob(globPattern)
	if err != nil {
		return err
	}
	outputFile, err := core.CreateFile(outputPath)
	defer outputFile.Close()

	if err != nil {
		return err
	}
	log.Printf("starting for %d files", len(files))
	missing := make(chan uint64)
	invalidFiles := make(chan string)

	var wg sync.WaitGroup

	run := core.MakeFileProcessor(func(filename string) error {
		defer wg.Done()
		reader, err := core.OpenFile(filename)
		if err != nil {
			invalidFiles <- filename
			return err
		}
		defer reader.Close()
		numbers, err := GetMissingBlockNumbers(reader, blockchain)
		if err != nil {
			invalidFiles <- filename
		}
		for _, number := range numbers {
			missing <- number
		}
		return err
	})

	var writeWg sync.WaitGroup
	go func() {
		defer writeWg.Done()
		for blockNumber := range missing {
			fmt.Fprintf(outputFile, "{\"block\": %d}\n", blockNumber)
		}
	}()
	go func() {
		defer writeWg.Done()
		for filename := range invalidFiles {
			fmt.Fprintf(outputFile, "{\"file\": %s}\n", filename)
		}
	}()
	writeWg.Add(2)

	for _, filename := range files {
		wg.Add(1)
		go run(filename)
	}
	wg.Wait()

	close(missing)
	close(invalidFiles)
	writeWg.Wait()

	return nil
}
