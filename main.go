package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
)

var start = flag.Int("start", 55406491, "start ledger index")
var end = flag.Int("end", 1, "last ledger index")
var filepath = flag.String("filepath", "", "base file path")

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	if *filepath == "" {
		log.Fatal("filepath not given")
	}

	fetchXRPData(*filepath, *start, *end, interrupt)
}
