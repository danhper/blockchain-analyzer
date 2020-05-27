package eos

import (
	"testing"
	"time"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/stretchr/testify/assert"
)

func TestParseBlock(t *testing.T) {
	rawBlock := core.ReadAllBlocks("eos")[0]
	block, err := New().ParseBlock(rawBlock)

	assert.Nil(t, err)
	assert.Equal(t, uint64(120893628), block.Number())
	assert.Equal(t, 8, block.TransactionsCount())
	expectedTime := time.Date(2020, time.Month(5), 16, 0, 10, 43, 0, time.UTC)
	assert.Equal(t, expectedTime, block.Time())
}

func TestParseBlockWithoutTrx(t *testing.T) {
	rawBlock := core.ReadAllBlocks("eos")[3]
	block, err := New().ParseBlock(rawBlock)

	assert.Nil(t, err)
	assert.Equal(t, uint64(120893629), block.Number())
	assert.Equal(t, 10, block.TransactionsCount())
}

func TestListActions(t *testing.T) {
	rawBlock := core.ReadAllBlocks("eos")[0]
	block, _ := New().ParseBlock(rawBlock)
	actions := block.ListActions()
	assert.Len(t, actions, 176)
}
