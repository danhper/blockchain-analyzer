package tezos

import (
	"fmt"
	"net/http"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/danhper/blockchain-analyzer/fetcher"
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

type Content struct {
	Kind        string
	Source      string
	Destination string
	Amount      string
}

type Operation struct {
	Hash     string
	Contents []Content
}

type BlockHeader struct {
	Level           uint64
	Timestamp       string
	ParsedTimestamp time.Time
}

type Block struct {
	Header     BlockHeader
	Operations [][]Operation
	actions    []core.Action
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
	parsedTime, err := time.Parse(time.RFC3339, block.Header.Timestamp)
	if err != nil {
		return nil, err
	}
	block.Header.ParsedTimestamp = parsedTime
	return &block, nil
}

func (t *Tezos) EmptyBlock() core.Block {
	return &Block{}
}

func (b *Block) Number() uint64 {
	return b.Header.Level
}

func (b *Block) Time() time.Time {
	return b.Header.ParsedTimestamp
}

func (b *Block) TransactionsCount() int {
	total := 0
	for _, operations := range b.Operations {
		total += len(operations)
	}
	return total
}

func (b *Block) ListActions() []core.Action {
	if len(b.actions) > 0 {
		return b.actions
	}
	var result []core.Action
	for _, operations := range b.Operations {
		for _, operation := range operations {
			for _, content := range operation.Contents {
				result = append(result, content)
			}
		}
	}
	b.actions = result
	return result
}

func (c Content) Name() string {
	return c.Kind
}

func (c Content) Receiver() string {
	return c.Destination
}

func (c Content) Sender() string {
	return c.Source
}
