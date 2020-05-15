package core

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
)

func GetFixturesPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(filename), "fixtures")
}

func GetFixture(filename string) string {
	return path.Join(GetFixturesPath(), filename)
}

func GetXRPLedgersReader() io.ReadCloser {
	reader, err := os.Open(GetFixture("xrp-ledgers-54387273--54387372.jsonl.gz"))
	if err != nil {
		panic(err)
	}
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		panic(err)
	}
	return gzipReader
}

func ReadAllXRPRawLedgers() [][]byte {
	reader := GetXRPLedgersReader()
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return bytes.Split(content, []byte{'\n'})
}
