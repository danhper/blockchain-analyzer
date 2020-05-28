package processor

import (
	"log"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/ugorji/go/codec"
)

func ExportToMsgpack(
	blockchain core.Blockchain,
	globPattern string,
	start, end uint64,
	outputDir string,
) error {
	files, err := filepath.Glob(globPattern)
	if err != nil {
		return err
	}
	processed := 0
	fileDone := make(chan bool)
	var wg sync.WaitGroup

	exportFile := core.MakeFileProcessor(func(filename string) error {
		defer wg.Done()
		reader, err := core.OpenFile(filename)
		if err != nil {
			return err
		}
		defer reader.Close()

		outputFilename := strings.Replace(path.Base(filename), "jsonl", "dat", 1)
		outputFilepath := path.Join(outputDir, outputFilename)
		writer, err := core.CreateFile(outputFilepath)
		if err != nil {
			return err
		}
		defer writer.Close()

		for block := range YieldBlocks(reader, blockchain, JSONFormat) {
			if (start == 0 || block.Number() >= start) &&
				(end == 0 || block.Number() <= end) {
				var rawBlock []byte
				enc := codec.NewEncoderBytes(&rawBlock, msgpackHandle)
				if err := enc.Encode(block); err != nil {
					return err
				}
				writer.Write(rawBlock)
			}
		}
		fileDone <- true
		return err
	})

	go func() {
		for range fileDone {
			processed++
			log.Printf("files processed: %d/%d", processed, len(files))
		}
	}()

	log.Printf("exporting %d files", len(files))

	for _, filename := range files {
		wg.Add(1)
		go exportFile(filename)
	}

	wg.Wait()
	close(fileDone)

	return nil
}
