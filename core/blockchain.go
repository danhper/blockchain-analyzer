package core

type Blockchain interface {
	FetchData(filepath string, start, end uint64) error
}
