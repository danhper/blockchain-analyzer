package processor

import (
	"io"
	"testing"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/danhper/blockchain-analyzer/xrp"
	"github.com/stretchr/testify/assert"
)

func computeBlockNumbers(reader io.Reader, blockchain core.Blockchain, start, end uint64) *core.MissingBlocks {
	missingBlocks := core.NewMissingBlocks(start, end)
	for block := range YieldBlocks(reader, blockchain, JSONFormat) {
		missingBlocks.AddBlock(block)
	}
	return missingBlocks
}

func TestComputeBlockNumbers(t *testing.T) {
	reader := core.GetFixtureReader(core.XRPValidLedgersFilename)
	blockchain := xrp.New()
	blocks := computeBlockNumbers(reader, blockchain, 0, 0)
	assert.Len(t, blocks.Seen, 100)
	assert.Contains(t, blocks.Seen, uint64(54387329))
}

func TestGetMissingBlockNumbersValid(t *testing.T) {
	reader := core.GetFixtureReader(core.XRPValidLedgersFilename)
	blockchain := xrp.New()
	blockNumbers := computeBlockNumbers(reader, blockchain, 54387321, 54387329)
	missing := blockNumbers.Result()
	assert.Len(t, missing, 0)
}

func TestGetMissingBlockNumbersInvalid(t *testing.T) {
	reader := core.GetFixtureReader(core.XRPMissingLedgersFilename)
	blockchain := xrp.New()
	missingBlockNumbers := computeBlockNumbers(reader, blockchain, 123, 126)
	missing := missingBlockNumbers.Compute()
	assert.Len(t, missing, 1)
	assert.Equal(t, missing[0], uint64(124))
}

func TestCountTransactions(t *testing.T) {
	blockchain := xrp.New()
	filepath := core.GetFixture(core.XRPValidLedgersFilename)
	count, err := CountTransactions(blockchain, filepath, uint64(0), uint64(0))
	assert.Nil(t, err)
	assert.Equal(t, 4518, count)
}

func TestYieldAllDuplicated(t *testing.T) {
	blockchain := xrp.New()
	fixtures := core.GetFixture(core.XRPDuplicatedLedgersFilename)
	blocksChan, err := YieldAllBlocks(fixtures, blockchain, uint64(0), uint64(0))
	assert.Nil(t, err)
	var blocks []core.Block
	for block := range blocksChan {
		blocks = append(blocks, block)
	}
	assert.Equal(t, 3, len(blocks))
}

func TestCountActionsOverTime(t *testing.T) {
	blockchain := xrp.New()
	filepath := core.GetFixture(core.XRPValidLedgersFilename)
	actionsCount, err := CountActionsOverTime(
		blockchain, filepath, uint64(0), uint64(0), time.Minute, core.ActionName)
	assert.Nil(t, err)
	assert.Len(t, actionsCount.Actions, 7)
	lastGroup := time.Date(2020, 3, 27, 20, 55, 0, 0, time.UTC)
	assert.Contains(t, actionsCount.Actions, lastGroup)
	assert.Equal(t, uint64(96), actionsCount.Actions[lastGroup].GetCount("Payment"))
	beforeLastGroup := time.Date(2020, 3, 27, 20, 54, 0, 0, time.UTC)
	assert.Contains(t, actionsCount.Actions, beforeLastGroup)
	assert.Equal(t, uint64(519), actionsCount.Actions[beforeLastGroup].GetCount("OfferCreate"))
}

func TestCountTransactionsOverTime(t *testing.T) {
	blockchain := xrp.New()
	filepath := core.GetFixture(core.XRPValidLedgersFilename)
	actionsCount, err := CountTransactionsOverTime(
		blockchain, filepath, uint64(0), uint64(0), time.Minute)
	assert.Nil(t, err)
	assert.Len(t, actionsCount.TransactionCounts, 7)
	lastGroup := time.Date(2020, 3, 27, 20, 55, 0, 0, time.UTC)
	assert.Equal(t, 451, actionsCount.TransactionCounts[lastGroup])
	beforeLastGroup := time.Date(2020, 3, 27, 20, 54, 0, 0, time.UTC)
	assert.Equal(t, 803, actionsCount.TransactionCounts[beforeLastGroup])
}

func TestGroupActions(t *testing.T) {
	blockchain := xrp.New()
	filepath := core.GetFixture(core.XRPValidLedgersFilename)
	actionsCount, err := GroupActions(
		blockchain, filepath, uint64(0), uint64(0), core.ActionName, false)
	assert.Nil(t, err)
	assert.Equal(t, uint64(1129), actionsCount.GetCount("Payment"))
	assert.Equal(t, uint64(3088), actionsCount.GetCount("OfferCreate"))
}
