package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const producerURL string = "https://api.main.alohaeos.com:443"

func fetchBlockWithRetry(client *http.Client, blockNumber, retries int) (result []byte, err error) {
	url := fmt.Sprintf("%s/v1/chain/get_block/%d", producerURL, blockNumber)
	resp, err := client.Get(url)
	if err == nil {
		result, err = ioutil.ReadAll(resp.Body)
	}
	if err != nil && retries > 0 {
		return fetchBlockWithRetry(client, blockNumber, retries-1)
	}
	return
}

func fetchBlock(blockNumber int, writer io.Writer, client *http.Client) error {
	result, err := fetchBlockWithRetry(client, blockNumber, 3)
	if err != nil {
		return err
	}
	result = append(result, '\n')
	_, err = writer.Write(result)
	return err
}

type EOSContext struct {
	doneCount  int
	totalCount int
}

func NewEOSContext(totalCount int) *EOSContext {
	return &EOSContext{
		doneCount:  0,
		totalCount: totalCount,
	}
}

func fetchBatch(filepath string, start, end int, context *EOSContext) error {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	gzipFile, err := openGZFile(makeFilename(filepath, start, end))
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	for block := end; block >= start; block-- {
		if context.doneCount%100 == 0 {
			log.Printf("%d/%d", context.doneCount, context.totalCount)
		}
		err := fetchBlock(block, gzipFile, client)
		if err != nil {
			log.Printf("could not fetch block %d: %s", block, err.Error())
		}
		context.doneCount++
	}

	return nil
}

func fetchEOSData(filepath string, start, end int) error {
	totalCount := end - start + 1
	context := NewEOSContext(totalCount)
	log.Printf("fetching %d blocks", totalCount)
	for block := end; block >= start; block -= batchSize {
		currentFirst := block - batchSize + 1
		if err := fetchBatch(filepath, currentFirst, block, context); err != nil {
			return err
		}
	}
	return nil
}
