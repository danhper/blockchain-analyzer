package xrp

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/danhper/blockchain-data-fetcher/core"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type XRP struct {
}

func New() *XRP {
	return &XRP{}
}

type Transaction struct {
	Account string
}

type Ledger struct {
	Index        uint64
	Transactions []Transaction
}

type XRPLedgerResponse struct {
	Result struct {
		LedgerIndex uint64 `json:"ledger_index"`
		Ledger      Ledger
	}
}

func ParseRawLedger(rawLedger []byte) (*Ledger, error) {
	var response XRPLedgerResponse
	if err := json.Unmarshal(rawLedger, &response); err != nil {
		return nil, err
	}
	response.Result.Ledger.Index = response.Result.LedgerIndex
	return &response.Result.Ledger, nil
}

func (e *XRP) ParseBlock(rawLine []byte) (core.Block, error) {
	return ParseRawLedger(rawLine)
}

func (e *XRP) FetchData(filepath string, start, end uint64) error {
	return fetchXRPData(filepath, start, end)
}

func (l *Ledger) Number() uint64 {
	return l.Index
}

func (l *Ledger) TransactionsCount() int {
	return len(l.Transactions)
}
