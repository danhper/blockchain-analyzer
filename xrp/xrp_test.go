package xrp

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getDataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(filename), "data", "xrp-ledgers-54387273--54387372.jsonl.gz")
}

func getRawLedgers() [][]byte {
	reader, err := os.Open(getDataPath())
	if err != nil {
		panic(err)
	}
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		panic(err)
	}
	return bytes.Split(content, []byte{'\n'})
}

func TestParseRawLedger(t *testing.T) {
	rawLedgers := getRawLedgers()
	rawLedger := rawLedgers[0]
	ledger, err := ParseRawLedger(rawLedger)

	assert.Nil(t, err)
	assert.Equal(t, ledger.Number(), uint64(54387329))
}
