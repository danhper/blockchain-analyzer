package eos

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/danhper/blockchain-analyzer/fetcher"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	defaultProducerURL string = "https://api.main.alohaeos.com:443"
	timeLayout         string = "2006-01-02T15:04:05.999"
)

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

type MapOrString map[string]interface{}

func (t *MapOrString) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var id string
		if err := json.Unmarshal(b, &id); err != nil {
			return err
		}
		*t = MapOrString(make(map[string]interface{}))
		return nil
	} else if b[0] == '{' {
		return json.Unmarshal(b, (*map[string]interface{})(t))
	}
	return fmt.Errorf("expected '{' or '\"', got %s", string(b))
}

type Action struct {
	Account       string
	Name          string
	Authorization []struct {
		Actor      string
		Permission string
	}
	Data MapOrString
}

type Transaction struct {
	Actions     []Action
	Expiration  string
	RefBlockNum int `json:"ref_block_num"`
}

type Trx struct {
	Id          string
	Signatures  []string
	Transaction Transaction
}

type TrxOrString Trx

func (t *TrxOrString) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var id string
		if err := json.Unmarshal(b, &id); err != nil {
			return err
		}
		*t = TrxOrString{Id: id}
		return nil
	} else if b[0] == '{' {
		return json.Unmarshal(b, (*Trx)(t))
	}
	return fmt.Errorf("expected '{' or '\"', got %s", string(b))
}

type FullTransaction struct {
	Status string
	Trx    TrxOrString
}

type Block struct {
	BlockNumber  uint64 `json:"block_num"`
	Timestamp    string
	parsedTime   time.Time
	Transactions []FullTransaction
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

func (e *EOS) ParseBlock(rawBlock []byte) (core.Block, error) {
	var block Block
	if err := json.Unmarshal(rawBlock, &block); err != nil {
		return nil, err
	}
	parsedTime, err := time.Parse(timeLayout, block.Timestamp)
	if err != nil {
		return nil, err
	}
	block.parsedTime = parsedTime
	return &block, json.Unmarshal(rawBlock, &block)
}

func (e *EOS) EmptyBlock() core.Block {
	return &Block{}
}

func (b *Block) Number() uint64 {
	return b.BlockNumber
}

func (b *Block) Time() time.Time {
	return b.parsedTime
}

func (b *Block) TransactionsCount() int {
	return len(b.Transactions)
}

func (b *Block) Actions() []Action {
	var actions []Action
	for _, transaction := range b.Transactions {
		for _, action := range transaction.Trx.Transaction.Actions {
			actions = append(actions, action)
		}
	}
	return actions
}

func (b *Block) GetActionsCount(prop core.ActionProperty) *core.ActionsCount {
	if prop != core.ActionName {
		panic(fmt.Errorf("action's %d not supported in XRP", prop))
	}
	actionsCount := core.NewActionsCount()
	for _, action := range b.Actions() {
		if prop == core.ActionName {
			actionsCount.Increment(action.Name)
		} else {
			actionsCount.Increment(action.Account)
		}
	}
	return actionsCount
}
