package eos

import (
	"encoding/json"

	"github.com/danhper/blockchain-data-fetcher/core"
)

type EOS struct{}

func (e *EOS) FetchData(filepath string, start, end uint64) error {
	return fetchEOSData(filepath, start, end)
}

type Block struct {
	BlockNumber uint64 `json:"block_num"`
}

func (b *Block) Number() uint64 {
	return b.BlockNumber
}

func New() *EOS {
	return &EOS{}
}

func (e *EOS) ParseBlock(rawLine []byte) (core.Block, error) {
	var block Block
	if err := json.Unmarshal(rawLine, &block); err != nil {
		return nil, err
	}
	return &block, nil
}
