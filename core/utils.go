package core

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
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

func CreateFile(name string) (io.WriteCloser, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(name, ".gz") {
		return gzip.NewWriter(file), nil
	}
	return file, nil
}

func OpenFile(name string) (io.ReadCloser, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(name, ".gz") {
		return gzip.NewReader(file)
	}
	return file, nil
}

func SortU64Slice(values []uint64) {
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
}

func MakeFileProcessor(f func(string) error) func(string) {
	return func(filename string) {
		log.Printf("processing %s", filename)
		if err := f(filename); err != nil {
			log.Printf("error while processing %s: %s", filename, err.Error())
		}
	}
}
