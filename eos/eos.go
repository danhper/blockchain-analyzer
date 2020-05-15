package eos

type EOS struct {
}

func (e *EOS) FetchData(filepath string, start, end uint64) error {
	return fetchEOSData(filepath, start, end)
}

func New() *EOS {
	return &EOS{}
}
