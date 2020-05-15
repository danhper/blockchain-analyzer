package processor

import (
	"testing"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/danhper/blockchain-data-fetcher/xrp"
	"github.com/stretchr/testify/assert"
)

func TestComputeBlockNumbers(t *testing.T) {
	reader := core.GetXRPLedgersReader()
	blockchain := xrp.New()
	blocks := ComputeBlockNumbers(reader, blockchain)
	assert.Len(t, blocks, 100)
	assert.Contains(t, blocks, uint64(54387329))
}
