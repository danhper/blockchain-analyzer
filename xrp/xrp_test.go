package xrp

import (
	"testing"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/stretchr/testify/assert"
)

func TestParseRawLedger(t *testing.T) {
	rawLedger := core.ReadAllXRPRawLedgers()[0]
	ledger, err := ParseRawLedger(rawLedger)

	assert.Nil(t, err)
	assert.Equal(t, ledger.Number(), uint64(54387329))
}
