package tezos

import (
	"fmt"
	"net/http"
	"os"

	jsoniter "github.com/json-iterator/go"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/danhper/blockchain-data-fetcher/fetcher"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const defaultRPCEndpoint string = "https://api.tezos.org.ua"

type Tezos struct {
	RPCEndpoint string
}

func (t *Tezos) makeRequest(client *http.Client, blockNumber uint64) (*http.Response, error) {
	url := fmt.Sprintf("%s/chains/main/blocks/%d", t.RPCEndpoint, blockNumber)
	return client.Get(url)
}

func (t *Tezos) FetchData(filepath string, start, end uint64) error {
	context := fetcher.NewHTTPContext(start, end, t.makeRequest)
	return fetcher.FetchHTTPData(filepath, context)
}

type Operation struct {
	Hash string
}

type Block struct {
	Header struct {
		Level uint64
	}
	Operations [][]Operation
}

func (b *Block) Number() uint64 {
	return b.Header.Level
}

func (b *Block) TransactionsCount() int {
	total := 0
	for _, operations := range b.Operations {
		total += len(operations)
	}
	return total
}

func New() *Tezos {
	rpcEndpoint := os.Getenv("TEZOS_RPC_ENDPOINT")
	if rpcEndpoint == "" {
		rpcEndpoint = defaultRPCEndpoint
	}

	return &Tezos{
		RPCEndpoint: rpcEndpoint,
	}
}

func (t *Tezos) ParseBlock(rawLine []byte) (core.Block, error) {
	var block Block
	if err := json.Unmarshal(rawLine, &block); err != nil {
		return nil, err
	}
	return &block, nil
}
