package eos

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/danhper/blockchain-data-fetcher/fetcher"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const defaultProducerURL string = "https://api.main.alohaeos.com:443"

type EOS struct {
	ProducerURL string
}

func (e *EOS) makeRequest(client *http.Client, blockNumber uint64) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/chain/get_block", e.ProducerURL)
	data := fmt.Sprintf("{\"block_num_or_id\": %d}", blockNumber)
	return client.Post(url, "application/json", strings.NewReader(data))
}

func (e *EOS) FetchData(filepath string, start, end uint64) error {
	context := fetcher.NewHTTPContext(start, end, e.makeRequest)
	return fetcher.FetchHTTPData(filepath, context)
}

type Transaction struct {
	Status string
}

type Block struct {
	BlockNumber  uint64 `json:"block_num"`
	Transactions []Transaction
}

func (b *Block) Number() uint64 {
	return b.BlockNumber
}

func (b *Block) TransactionsCount() int {
	return len(b.Transactions)
}

func New() *EOS {
	producerURL := os.Getenv("EOS_PRODUCER_URL")
	if producerURL == "" {
		producerURL = defaultProducerURL
	}

	return &EOS{
		ProducerURL: producerURL,
	}
}

func (e *EOS) ParseBlock(rawLine []byte) (core.Block, error) {
	var block Block
	if err := json.Unmarshal(rawLine, &block); err != nil {
		return nil, err
	}
	return &block, nil
}
