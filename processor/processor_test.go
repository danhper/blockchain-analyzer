package processor

import (
	"testing"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/danhper/blockchain-analyzer/xrp"
	"github.com/stretchr/testify/assert"
)

func TestComputeBlockNumbers(t *testing.T) {
	reader := core.GetFixtureReader(core.XRPValidLedgersFilename)
	blockchain := xrp.New()
	blocks := ComputeBlockNumbers(reader, blockchain)
	assert.Len(t, blocks, 100)
	assert.Contains(t, blocks, uint64(54387329))
}

func TestGetMissingBlockNumbersValid(t *testing.T) {
	reader := core.GetFixtureReader(core.XRPValidLedgersFilename)
	blockchain := xrp.New()
	blockNumbers := ComputeBlockNumbers(reader, blockchain)
	missing := ComputeMissingBlockNumbers(blockNumbers, 54387321, 54387329)
	assert.Len(t, missing, 0)
}

func TestGetMissingBlockNumbersInvalid(t *testing.T) {
	reader := core.GetFixtureReader(core.XRPMissingLedgersFilename)
	blockchain := xrp.New()
	blockNumbers := ComputeBlockNumbers(reader, blockchain)
	missing := ComputeMissingBlockNumbers(blockNumbers, 123, 126)
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

func TestCountActions(t *testing.T) {
	blockchain := xrp.New()
	filepath := core.GetFixture(core.XRPValidLedgersFilename)
	actionsCount, err := CountActions(blockchain, filepath, uint64(0), uint64(0))
	assert.Nil(t, err)
	assert.Equal(t, uint64(1129), actionsCount.Get("Payment"))
	assert.Equal(t, uint64(3088), actionsCount.Get("OfferCreate"))
}

func TestCountActionsPerTime(t *testing.T) {
	blockchain := xrp.New()
	filepath := core.GetFixture(core.XRPValidLedgersFilename)
	actionsCount, err := CountActionsPerTime(
		blockchain, filepath, uint64(0), uint64(0), time.Minute, core.ActionName)
	assert.Nil(t, err)
	assert.Len(t, actionsCount.Actions, 7)
	lastGroup := time.Date(2020, 3, 27, 20, 55, 0, 0, time.UTC)
	assert.Contains(t, actionsCount.Actions, lastGroup)
	assert.Equal(t, uint64(96), actionsCount.Actions[lastGroup].Get("Payment"))
	beforeLastGroup := time.Date(2020, 3, 27, 20, 54, 0, 0, time.UTC)
	assert.Contains(t, actionsCount.Actions, beforeLastGroup)
	assert.Equal(t, uint64(519), actionsCount.Actions[beforeLastGroup].Get("OfferCreate"))
}
