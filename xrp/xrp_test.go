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

func TestListActions(t *testing.T) {
	rawLedger := core.ReadAllBlocks("xrp")[0]
	ledger, _ := ParseRawLedger(rawLedger)
	actions := ledger.ListActions()
	assert.Len(t, actions, 33)
}
