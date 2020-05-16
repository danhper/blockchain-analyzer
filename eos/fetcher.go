package eos

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/danhper/blockchain-data-fetcher/core"
)

const producerURL string = "https://api.main.alohaeos.com:443"

func fetchBlockWithRetry(client *http.Client, blockNumber uint64, retries int) (result []byte, err error) {
	url := fmt.Sprintf("%s/v1/chain/get_block", producerURL)
	data := fmt.Sprintf("{\"block_num_or_id\": %d}", blockNumber)
	resp, err := client.Post(url, "application/json", strings.NewReader(data))
	if err == nil {
		result, err = ioutil.ReadAll(resp.Body)
	}
	if err != nil && retries > 0 {
		log.Printf("error: %s, retrying", err.Error())
		return fetchBlockWithRetry(client, blockNumber, retries-1)
	}
	return
}

func fetchBlock(blockNumber uint64, client *http.Client) ([]byte, error) {
	return fetchBlockWithRetry(client, blockNumber, 3)
}

type EOSContext struct {
	doneCount  uint64
	totalCount uint64
}

func NewEOSContext(totalCount uint64) *EOSContext {
	return &EOSContext{
		doneCount:  0,
		totalCount: totalCount,
	}
}

func fetchBlocks(blocks <-chan uint64, results chan<- []byte) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	for block := range blocks {
		result, err := fetchBlock(block, client)
		if err != nil {
			log.Printf("could not fetch block %d: %s", block, err.Error())
		}
		results <- result
	}
}

func fetchBatch(filepath string, start, end uint64, context *EOSContext) error {
	gzipFile, err := core.CreateFile(core.MakeFilename(filepath, start, end))
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	workersCount := 10
	blocksCount := end - start + 1
	jobs := make(chan uint64, blocksCount)
	results := make(chan []byte, blocksCount)

	for w := 1; w <= workersCount; w++ {
		go fetchBlocks(jobs, results)
	}

	for block := end; block >= start; block-- {
		jobs <- block
	}
	close(jobs)
	for i := uint64(0); i < blocksCount; i++ {
		result := <-results
		result = append(result, '\n')
		gzipFile.Write(result)

		context.doneCount++
		if context.doneCount%100 == 0 {
			log.Printf("%d/%d", context.doneCount, context.totalCount)
		}
	}

	return nil
}

func fetchEOSData(filepath string, start, end uint64) error {
	totalCount := end - start + 1
	context := NewEOSContext(totalCount)
	log.Printf("fetching %d blocks", totalCount)
	for block := end; block >= start; block -= core.BatchSize {
		currentFirst := block - core.BatchSize + 1
		if err := fetchBatch(filepath, currentFirst, block, context); err != nil {
			return err
		}
	}
	return nil
}
