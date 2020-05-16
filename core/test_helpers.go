package core

import (
	"bytes"
	"io"
	"io/ioutil"
	"path"
	"runtime"
)

const (
	RealLedgersFilename    string = "xrp-ledgers-54387273--54387372.jsonl.gz"
	MissingLedgersFilename string = "xrp-missing-block.jsonl"
)

func GetFixturesPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(filename), "fixtures")
}

func GetFixture(filename string) string {
	return path.Join(GetFixturesPath(), filename)
}

func GetXRPLedgersReader(filename string) io.ReadCloser {
	reader, err := OpenFile(GetFixture(filename))
	if err != nil {
		panic(err)
	}
	return reader
}

func ReadAllXRPRawLedgers() [][]byte {
	reader := GetXRPLedgersReader(RealLedgersFilename)
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return bytes.Split(content, []byte{'\n'})
}
