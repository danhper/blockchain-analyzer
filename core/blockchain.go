package core

import (
	"time"
)

type Blockchain interface {
	FetchData(filepath string, start, end uint64) error
	ParseBlock(rawLine []byte) (Block, error)
	EmptyBlock() Block
}

type Block interface {
	Number() uint64
	TransactionsCount() int
	Time() time.Time
	ListActions() []Action
}

type Action interface {
	Sender() string
	Receiver() string
	Name() string
}
