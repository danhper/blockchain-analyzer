package xrp

type XRP struct {
}

func (e *XRP) FetchData(filepath string, start, end uint64) error {
	return fetchXRPData(filepath, start, end)
}

func New() *XRP {
	return &XRP{}
}
