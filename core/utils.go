package core

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	BatchSize uint64 = 100000
)

func MakeFilename(filePath string, first, last uint64) string {
	splitted := strings.SplitN(filePath, ".", 2)
	return fmt.Sprintf("%s-%d--%d.%s", splitted[0], first, last, splitted[1])
}

func MakeErrFilename(filePath string, first, last uint64) string {
	splitted := strings.SplitN(filePath, ".", 2)
	return fmt.Sprintf("%s-%d--%d-errors.%s", splitted[0], first, last, splitted[1])
}

func OpenGZFile(name string) (io.WriteCloser, error) {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	gzipFile := gzip.NewWriter(file)
	return gzipFile, nil
}
