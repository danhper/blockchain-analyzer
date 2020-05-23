package tezos

import (
	"testing"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/stretchr/testify/assert"
)

func TestParseBlock(t *testing.T) {
	rawBlock := core.ReadAllBlocks("tezos")[0]
	block, err := New().ParseBlock(rawBlock)

	assert.Nil(t, err)
	assert.Equal(t, uint64(10000), block.Number())
	assert.Equal(t, 8, block.TransactionsCount())

	expectedTime := time.Date(2018, 7, 7, 17, 06, 27, 0, time.UTC)
	assert.Equal(t, expectedTime, block.Time())
}

func TestGetActionsCount(t *testing.T) {
	rawBlock := core.ReadAllBlocks("tezos")[1]
	block, _ := New().ParseBlock(rawBlock)
	assert.Equal(t, uint64(8), block.GetActionsCount().Get("endorsement"))
	assert.Equal(t, uint64(1), block.GetActionsCount().Get("delegation"))
}
