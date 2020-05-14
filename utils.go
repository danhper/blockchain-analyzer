package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	batchSize uint64 = 100000
)

func makeFilename(filePath string, first, last uint64) string {
	splitted := strings.SplitN(filePath, ".", 2)
	return fmt.Sprintf("%s-%d--%d.%s", splitted[0], first, last, splitted[1])
}

func makeErrFilename(filePath string, first, last uint64) string {
	splitted := strings.SplitN(filePath, ".", 2)
	return fmt.Sprintf("%s-%d--%d-errors.%s", splitted[0], first, last, splitted[1])
}

func openGZFile(name string) (io.WriteCloser, error) {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	gzipFile := gzip.NewWriter(file)
	return gzipFile, nil
}
