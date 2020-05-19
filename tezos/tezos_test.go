package tezos

import (
	"testing"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/stretchr/testify/assert"
)

func TestParseBlock(t *testing.T) {
	rawBlock := core.ReadAllBlocks("tezos")[0]
	block, err := New().ParseBlock(rawBlock)

	assert.Nil(t, err)
	assert.Equal(t, uint64(10000), block.Number())
	assert.Equal(t, 8, block.TransactionsCount())
}
