package xrp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/danhper/blockchain-data-fetcher/core"
)

const (
	wsURI    string = "wss://xrpl.ws"
	maxTries int    = 5
)

type WSError struct {
	message string
}

func (e *WSError) Error() string {
	return fmt.Sprintf("ws connection error: %s", e.message)
}

func makeMessage(ledgerIndex uint64) []byte {
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

func processWSMessages(
	conn *websocket.Conn, writer io.Writer, wg *sync.WaitGroup, quit chan struct{}, written chan<- uint64,
	abort chan<- error) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			abort <- &WSError{message: err.Error()}
			log.Println("read:", err)
			return
		}
		writer.Write(message)
		writer.Write([]byte{'\n'})

		ledger, err := ParseRawLedger(message)
		if err != nil {
			log.Printf("error while parsing message: %s", err.Error())
		}
		written <- ledger.Number()
		wg.Done()

		select {
		case <-quit:
			return
		case <-time.After(time.Millisecond):
		}
	}
}

func sendWSMessage(conn *websocket.Conn, ledger uint64) {
	message := makeMessage(ledger)
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Printf("could not fetch ledger %d: %s", ledger, err.Error())
	}
}

func writeFailed(filePath string, failed map[uint64]bool) error {
	// write failed blocks
	var errWriter io.Writer
	errWriter, err := core.OpenGZFile(filePath)
	if err != nil {
		log.Printf("could not open error file: %s", err.Error())
		return err
	}
	for ledger := range failed {
		errWriter.Write([]byte(fmt.Sprintf("%d\n", ledger)))
	}
	return nil
}

func fetchLedgersRange(start, end uint64, filePath string, context *XRPContext) (stop bool, err error) {
	toFetch := make(map[uint64]bool)
	for ledger := start; ledger <= end; ledger++ {
		toFetch[ledger] = true
	}
	writer, err := core.OpenGZFile(core.MakeFilename(filePath, start, end))
	if err != nil {
		return false, err
	}
	defer writer.Close()
	return fetchLedgersWithRetry(toFetch, writer, context)
}

func fetchLedgersWithRetry(toFetch map[uint64]bool, writer io.Writer, context *XRPContext) (stop bool, err error) {
	tries := 0
	for tries < maxTries {
		stop, err = fetchLedgers(toFetch, writer, context)
		if stop {
			return
		}

		// reconnect in case of websocket failures
		if err != nil && errors.Is(err, &WSError{}) {
			if err = context.Reconnect(); err != nil {
				log.Printf("error while reconnecting: %s\n", err.Error())
				return true, err
			}
		}
		if len(toFetch) == 0 {
			break
		}
		log.Printf("%d items left in batch, retrying", len(toFetch))
	}
	if len(toFetch) > 0 {
		writeFailed("failed.txt.gz", toFetch)
	}

	return
}

func fetchLedgers(toFetch map[uint64]bool, writer io.Writer, context *XRPContext) (bool, error) {
	var wg sync.WaitGroup
	quit := make(chan struct{})
	waiting := 0
	bufferSize := 20
	written := make(chan uint64, bufferSize)
	abort := make(chan error, 1)

	go processWSMessages(context.conn, writer, &wg, quit, written, abort)

	shouldWait := true

	defer func() {
		if shouldWait {
			wg.Wait()
			close(written)
			for index := range written {
				delete(toFetch, index)
			}
		}
		close(quit)
	}()

	ledgersToFetch := make([]uint64, len(toFetch))
	index := 0
	for ledger := range toFetch {
		ledgersToFetch[index] = ledger
		index++
	}
	index = 0
	for index < len(ledgersToFetch) {
		if waiting < bufferSize {
			ledger := ledgersToFetch[index]
			sendWSMessage(context.conn, ledger)
			waiting++
			wg.Add(1)
			if context.doneCount%100 == 0 {
				log.Printf("%d/%d", context.doneCount, context.totalCount)
			}
			context.doneCount++
			index++
		}
		select {
		case err := <-abort:
			shouldWait = false
			return false, err
		case <-context.interrupt:
			return true, nil
		case index := <-written:
			waiting--
			delete(toFetch, index)
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
	totalCount uint64
}

func NewXRPContext(interrupt chan os.Signal, totalCount uint64) (*XRPContext, error) {
	context := &XRPContext{
		interrupt:  interrupt,
		doneCount:  0,
		totalCount: totalCount,
	}
	return context, context.Reconnect()
}

func (c *XRPContext) Reconnect() (err error) {
	c.conn, _, err = websocket.DefaultDialer.Dial(wsURI, nil)
	return
}

func (c *XRPContext) Cleanup() error {
	if c.conn != nil {
		closeConnection(c.conn)
		return c.conn.Close()
	}
	return nil
}

func fetchXRPData(filepath string, start, end uint64) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	totalCount := end - start + 1
	context, err := NewXRPContext(interrupt, totalCount)
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Printf("fetching %d ledgers", totalCount)

	defer context.Cleanup()

	for ledger := end; ledger >= start; ledger -= core.BatchSize {
		currentStart := ledger - core.BatchSize + 1
		interrupted, err := fetchLedgersRange(currentStart, ledger, filepath, context)
		if err != nil || interrupted {
			break
		}
	}

	return nil
}
