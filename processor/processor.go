package processor

import (
	"bufio"
	"bytes"
	"io"
	"log"

	"github.com/danhper/blockchain-data-fetcher/core"
)

func ComputeBlockNumbers(reader io.Reader, blockchain core.Blockchain) map[uint64]bool {
	stream := bufio.NewReader(reader)
	result := make(map[uint64]bool)
	for i := 0; ; i++ {
		rawLine, err := stream.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("failed to read line %s\n", err.Error())
			continue
		}
		rawLine = bytes.ToValidUTF8(rawLine, []byte{})
		block, err := blockchain.ParseBlock(rawLine)
		if err != nil {
			log.Printf("could not parse: %s, line: %s", err.Error(), rawLine)
			continue
		}
		result[block.Number()] = true
	}
	return result
}

// func CheckMissingBlocks(reader io.Reader, blockchain Blockchain) []uint64 {
// 	return nil
// }
