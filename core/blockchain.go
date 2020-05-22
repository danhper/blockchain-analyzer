package core

type Blockchain interface {
	FetchData(filepath string, start, end uint64) error
	ParseBlock(rawLine []byte) (Block, error)
}

type Block interface {
	Number() uint64
	TransactionsCount() int
	GetActionsCount() *ActionsCount
}
