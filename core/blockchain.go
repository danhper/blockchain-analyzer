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
	GetActionsCount(ActionProperty) *ActionsCount
	Time() time.Time
}
