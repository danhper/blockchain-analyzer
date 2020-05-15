package xrp

import (
	"encoding/json"

	"github.com/danhper/blockchain-data-fetcher/core"
)

type XRP struct {
}

func New() *XRP {
	return &XRP{}
}

type Ledger struct {
	Index uint64
}

type XRPLedgerResponse struct {
	Result struct {
		LedgerIndex uint64 `json:"ledger_index"`
	}
}

func ParseRawLedger(rawLedger []byte) (*Ledger, error) {
	var response XRPLedgerResponse
	if err := json.Unmarshal(rawLedger, &response); err != nil {
		return nil, err
	}
	ledgerIndex := response.Result.LedgerIndex
	return &Ledger{Index: ledgerIndex}, nil
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
