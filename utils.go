package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	batchSize int = 100
)

func makeFilename(filePath string, first, last int) string {
	splitted := strings.SplitN(filePath, ".", 2)
	return fmt.Sprintf("%s-%d--%d.%s", splitted[0], first, last, splitted[1])
}

func openGZFile(name string) (io.WriteCloser, error) {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	gzipFile := gzip.NewWriter(file)
	return gzipFile, nil
}
