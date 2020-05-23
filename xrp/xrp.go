package xrp

import (
	"time"

	jsoniter "github.com/json-iterator/go"

	"github.com/danhper/blockchain-analyzer/core"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const rippleEpochOffset int64 = 946684800

type XRP struct {
}

func New() *XRP {
	return &XRP{}
}

type Transaction struct {
	Account         string
	TransactionType string
}

type Ledger struct {
	Index           uint64 `json:"-"`
	CloseTimestamp  int64  `json:"close_time"`
	parsedCloseTime time.Time
	Transactions    []Transaction
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
	ledger := response.Result.Ledger
	ledger.parsedCloseTime = time.Unix(ledger.CloseTimestamp+rippleEpochOffset, 0).UTC()
	ledger.Index = response.Result.LedgerIndex
	return &ledger, nil
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

func (l *Ledger) Time() time.Time {
	return l.parsedCloseTime
}

func (l *Ledger) TransactionsCount() int {
	return len(l.Transactions)
}

func (l *Ledger) GetActionsCount(prop core.ActionProperty) *core.ActionsCount {
	actionsCount := core.NewActionsCount()
	for _, transaction := range l.Transactions {
		actionsCount.Increment(transaction.TransactionType)
	}
	return actionsCount
}
