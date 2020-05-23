package core

import (
	"bytes"
	"io"
	"io/ioutil"
	"path"
	"runtime"
)

const (
	XRPValidLedgersFilename       string = "xrp-ledgers-54387273--54387372.jsonl.gz"
	XRPSimpleValidLedgersFilename string = "xrp-ledgers-simple-format-50287874--50287973.jsonl.gz"
	XRPMissingLedgersFilename     string = "xrp-missing-block.jsonl"
	XRPDuplicatedLedgersFilename  string = "xrp-duplicated.jsonl"

	EOSValidBlocksFilename   string = "eos-blocks-120893532--120893631.jsonl.gz"
	TezosValidBlocksFilename string = "tezos-blocks.jsonl"
)

func GetFixturesPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(filename), "fixtures")
}

func GetFixture(filename string) string {
	return path.Join(GetFixturesPath(), filename)
}

func GetFixtureReader(filename string) io.ReadCloser {
	reader, err := OpenFile(GetFixture(filename))
	if err != nil {
		panic(err)
	}
	return reader
}

func ReadAllBlocks(blockchainName string) [][]byte {
	var filename string
	switch blockchainName {
	case "eos":
		filename = EOSValidBlocksFilename
	case "xrp":
		filename = XRPValidLedgersFilename
	case "tezos":
		filename = TezosValidBlocksFilename
	default:
		panic("invalid blockchain: " + blockchainName)
	}
	reader := GetFixtureReader(filename)
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return bytes.Split(content, []byte{'\n'})
}

func ReadAllEOSBlocks() [][]byte {
	reader := GetFixtureReader(EOSValidBlocksFilename)
	defer reader.Close()
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	return bytes.Split(content, []byte{'\n'})
}
