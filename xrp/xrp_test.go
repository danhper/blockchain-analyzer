package xrp

import (
	"testing"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/stretchr/testify/assert"
)

func TestParseRawLedger(t *testing.T) {
	rawLedger := core.ReadAllBlocks("xrp")[0]
	ledger, err := ParseRawLedger(rawLedger)

	assert.Nil(t, err)
	assert.Equal(t, uint64(54387329), ledger.Number())
	assert.Equal(t, 33, ledger.TransactionsCount())
	expectedTime := time.Date(2020, 3, 27, 20, 52, 50, 0, time.UTC)
	assert.Equal(t, expectedTime, ledger.Time())
}

func TestGetActionsCount(t *testing.T) {
	rawLedger := core.ReadAllBlocks("xrp")[0]
	ledger, _ := ParseRawLedger(rawLedger)
	actionsCount := ledger.GetActionsCount(core.ActionName)
	assert.Equal(t, uint64(7), actionsCount.Get("Payment"))
	assert.Equal(t, uint64(25), actionsCount.Get("OfferCreate"))
	assert.Equal(t, uint64(1), actionsCount.Get("OfferCancel"))
}
