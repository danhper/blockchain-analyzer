package fetcher

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
)

type RequestSender func(*http.Client, uint64) (*http.Response, error)

func fetchBlockWithRetry(
	client *http.Client, context *HTTPContext,
	blockNumber uint64, retries int,
) (result []byte, err error) {
	resp, err := context.MakeRequest(client, blockNumber)
	if err == nil && resp.StatusCode == 200 {
		result, err = ioutil.ReadAll(resp.Body)
	}
	if (err != nil || resp.StatusCode != 200) && retries > 0 {
		log.Printf("error: %s (status %d), retrying", err.Error(), resp.StatusCode)
		time.Sleep(time.Second)
		return fetchBlockWithRetry(client, context, blockNumber, retries-1)
	}
	return
}

func fetchBlock(blockNumber uint64, client *http.Client, context *HTTPContext) ([]byte, error) {
	return fetchBlockWithRetry(client, context, blockNumber, 3)
}

type HTTPContext struct {
	DoneCount   uint64
	Start       uint64
	End         uint64
	MakeRequest RequestSender
}

func NewHTTPContext(start, end uint64, makeRequest RequestSender) *HTTPContext {
	return &HTTPContext{
		DoneCount:   0,
		Start:       start,
		End:         end,
		MakeRequest: makeRequest,
	}
}

func (c *HTTPContext) TotalCount() uint64 {
	return c.End - c.Start + 1
}

func fetchBlocks(context *HTTPContext, blocks <-chan uint64, results chan<- []byte) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	for block := range blocks {
		result, err := fetchBlock(block, client, context)
		if err != nil {
			log.Printf("could not fetch block %d: %s", block, err.Error())
		}
		results <- result
	}
}

func fetchBatch(filepath string, start, end uint64, context *HTTPContext) error {
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
		go fetchBlocks(context, jobs, results)
	}

	for block := end; block >= start; block-- {
		jobs <- block
	}
	close(jobs)
	for i := uint64(0); i < blocksCount; i++ {
		result := <-results
		result = append(bytes.TrimSpace(result), '\n')
		gzipFile.Write(result)

		context.DoneCount++
		if context.DoneCount%100 == 0 {
			log.Printf("%d/%d", context.DoneCount, context.TotalCount())
		}
	}

	return nil
}

func FetchHTTPData(filepath string, context *HTTPContext) error {
	log.Printf("fetching %d blocks", context.TotalCount())
	getNext := func(num uint64) uint64 {
		if num >= core.BatchSize {
			return num - core.BatchSize
		} else {
			return context.Start
		}
	}
	for block := context.End; block > context.Start; block = getNext(block) {
		var currentFirst uint64
		if block+1 < core.BatchSize || block+1-core.BatchSize < context.Start {
			currentFirst = context.Start
		} else {
			currentFirst = block + 1 - core.BatchSize
		}
		if err := fetchBatch(filepath, currentFirst, block, context); err != nil {
			return err
		}
	}
	return nil
}
