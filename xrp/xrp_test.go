package xrp

import (
	"bufio"
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

func TestParseRawLedgerSimpleFormat(t *testing.T) {
	reader := core.GetFixtureReader(core.XRPSimpleValidLedgersFilename)
	defer reader.Close()
	rawLedger, err := bufio.NewReader(reader).ReadBytes('\n')
	assert.Nil(t, err)
	ledger, err := ParseRawLedger(rawLedger)
	assert.Nil(t, err)
	assert.Equal(t, uint64(50387844), ledger.Number())
}

func TestGetActionsCount(t *testing.T) {
	rawLedger := core.ReadAllBlocks("xrp")[0]
	ledger, _ := ParseRawLedger(rawLedger)
	actionsCount := ledger.GetActionsCount(core.ActionName)
	assert.Equal(t, uint64(7), actionsCount.Get("Payment"))
	assert.Equal(t, uint64(25), actionsCount.Get("OfferCreate"))
	assert.Equal(t, uint64(1), actionsCount.Get("OfferCancel"))
}
