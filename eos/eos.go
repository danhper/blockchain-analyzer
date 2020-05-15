package eos

import "github.com/danhper/blockchain-data-fetcher/core"

type EOS struct {
}

func (e *EOS) FetchData(filepath string, start, end uint64) error {
	return fetchEOSData(filepath, start, end)
}

type Block struct {
	BlockNumber uint64
}

func (b *Block) Number() uint64 {
	return b.BlockNumber
}

func New() *EOS {
	return &EOS{}
}

func (e *EOS) ParseBlock(rawLine []byte) (core.Block, error) {
	return &Block{BlockNumber: 0}, nil
}
