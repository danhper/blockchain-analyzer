package eos

import (
	"testing"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/stretchr/testify/assert"
)

func TestParseBlock(t *testing.T) {
	rawBlock := core.ReadAllBlocks("eos")[0]
	block, err := New().ParseBlock(rawBlock)

	assert.Nil(t, err)
	assert.Equal(t, uint64(120893628), block.Number())
	assert.Equal(t, 8, block.TransactionsCount())
}

func TestParseBlockWithoutTrx(t *testing.T) {
	rawBlock := core.ReadAllBlocks("eos")[3]
	block, err := New().ParseBlock(rawBlock)

	assert.Nil(t, err)
	assert.Equal(t, uint64(120893629), block.Number())
	assert.Equal(t, 10, block.TransactionsCount())
}

func TestActionsCount(t *testing.T) {
	rawBlock := core.ReadAllBlocks("eos")[0]
	block, _ := New().ParseBlock(rawBlock)
	actionsCount := block.GetActionsCount()
	assert.Equal(t, uint64(170), actionsCount.Get("transfer"))
	assert.Equal(t, uint64(1), actionsCount.Get("updaterating"))
	assert.Equal(t, uint64(1), actionsCount.Get("write"))
	assert.Equal(t, uint64(1), actionsCount.Get("clearing"))
	assert.Equal(t, uint64(1), actionsCount.Get("reveal"))
	assert.Equal(t, uint64(1), actionsCount.Get("cancelorder"))
}
