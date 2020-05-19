package eos

import (
	"testing"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/stretchr/testify/assert"
)

func TestParseBlock(t *testing.T) {
	rawBlock := core.ReadAllBlocks("eos")[0]
	ledger, err := New().ParseBlock(rawBlock)

	assert.Nil(t, err)
	assert.Equal(t, uint64(120893628), ledger.Number())
	assert.Equal(t, 8, ledger.TransactionsCount())
}
