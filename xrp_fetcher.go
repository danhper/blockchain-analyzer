package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsURI string = "wss://xrpl.ws"
)

func makeMessage(ledgerIndex int) []byte {
	params := map[string]interface{}{
		"command":      "ledger",
		"ledger_index": ledgerIndex,
		"transactions": true,
		"expand":       true,
	}
	message, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	return message
}

func processWSMessages(conn *websocket.Conn, writer io.Writer, wg *sync.WaitGroup, quit chan struct{}, written chan<- bool) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		writer.Write(message)
		writer.Write([]byte{'\n'})
		wg.Done()
		written <- true
		select {
		case <-quit:
			return
		case <-time.After(time.Millisecond):
		}
	}
}

func sendWSMessage(conn *websocket.Conn, ledger int) {
	if err := conn.WriteMessage(websocket.TextMessage, makeMessage(ledger)); err != nil {
		log.Printf("could not fetch ledger %d: %s", ledger, err.Error())
	}
}

func fetchLedgers(start, end int, filePath string, context *XRPContext) (bool, error) {
	writer, err := openGZFile(makeFilename(filePath, start, end))
	if err != nil {
		return false, err
	}
	defer writer.Close()

	var wg sync.WaitGroup
	quit := make(chan struct{})
	waiting := 0
	bufferSize := 20
	written := make(chan bool, bufferSize)

	go processWSMessages(context.conn, writer, &wg, quit, written)

	defer func() {
		wg.Wait()
		close(quit)
	}()

	for ledger := start; ledger <= end; {
		if waiting < bufferSize {
			sendWSMessage(context.conn, ledger)
			waiting++
			wg.Add(1)
			if context.doneCount%100 == 0 {
				log.Printf("%d/%d", context.doneCount, context.totalCount)
			}
			context.doneCount++
			ledger++
		}
		select {
		case <-context.interrupt:
			return true, nil
		case <-written:
			waiting--
		case <-time.After(time.Millisecond):
		}
	}
	return false, nil
}

func closeConnection(conn *websocket.Conn) {
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := conn.WriteMessage(websocket.CloseMessage, msg)
	if err != nil {
		log.Println("write close:", err)
		return
	}
	done := make(chan struct{})
	go func() {
		conn.ReadMessage()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
	}
}

type XRPContext struct {
	conn       *websocket.Conn
	interrupt  chan os.Signal
	doneCount  int
	totalCount int
}

func fetchXRPData(filepath string, start, end int) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn, _, err := websocket.DefaultDialer.Dial(wsURI, nil)
	if err != nil {
		return err
	}

	defer func() {
		closeConnection(conn)
		conn.Close()
	}()

	totalCount := end - start + 1
	log.Printf("fetching %d ledgers", totalCount)
	context := &XRPContext{conn: conn, interrupt: interrupt, totalCount: totalCount}

	for ledger := end; ledger >= start; ledger -= batchSize {
		currentStart := ledger - batchSize + 1
		interrupted, err := fetchLedgers(currentStart, ledger, filepath, context)
		if err != nil || interrupted {
			break
		}
	}

	return nil
}
