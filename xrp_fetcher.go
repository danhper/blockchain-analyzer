package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsURI     string = "wss://xrpl.ws"
	batchSize int    = 100000
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

func openGZFile(name string) (io.WriteCloser, error) {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	gzipFile := gzip.NewWriter(file)
	return gzipFile, nil
}

type Context struct {
	stop       chan struct{}
	done       chan struct{}
	conn       *websocket.Conn
	totalCount int
}

func NewContext(conn *websocket.Conn, totalCount int) *Context {
	return &Context{
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
		conn:       conn,
		totalCount: totalCount,
	}
}

type LoopContext struct {
	written chan bool
	stop    chan struct{}
}

func NewLoopContext() *LoopContext {
	return &LoopContext{
		written: make(chan bool),
		stop:    make(chan struct{}),
	}
}

func processWSMessages(conn *websocket.Conn, writer io.Writer, context *LoopContext) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		writer.Write(message)
		writer.Write([]byte{'\n'})
		context.written <- true
		select {
		case <-context.stop:
			return
		case <-time.After(time.Millisecond):
		}
	}
}

func makeFilename(filePath string, firstLedger, lastLedger int) string {
	splitted := strings.SplitN(filePath, ".", 2)
	return fmt.Sprintf("%s-%d--%d.%s", splitted[0], firstLedger, lastLedger, splitted[1])
}

func fetchLedgers(firstLedger, lastLedger int, filePath string, context *Context) error {
	doneCount := 0
	totalToFetch := lastLedger - firstLedger + 1
	pending := 0
	currentLedger := lastLedger
	loopContext := NewLoopContext()
	writer, err := openGZFile(makeFilename(filePath, firstLedger, lastLedger))
	if err != nil {
		return err
	}
	defer writer.Close()

	defer func() {
		close(loopContext.stop)
		for {
			select {
			case <-loopContext.written:
			case <-time.After(time.Second):
				close(context.done)
				return
			}
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	go processWSMessages(context.conn, writer, loopContext)

	for {
		select {
		case <-context.stop:
			return nil
		case <-loopContext.written:
			pending--
			doneCount++
			if doneCount%100 == 0 {
				log.Printf("%d/%d\n", doneCount, context.totalCount)
			}
			if doneCount == totalToFetch {
				return nil
			}
		case <-ticker.C:
			if pending < 20 {
				if err = context.conn.WriteMessage(websocket.TextMessage, makeMessage(currentLedger)); err != nil {
					return err
				}
				currentLedger--
				pending++
			}
		}
	}
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

func fetchXRPData(filepath string, startLedger, endLedger int, interrupt chan os.Signal) {
	conn, _, err := websocket.DefaultDialer.Dial(wsURI, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	defer func() {
		closeConnection(conn)
		conn.Close()
	}()

	totalCount := startLedger - endLedger + 1
	context := NewContext(conn, totalCount)
	running := true

	for ledger := startLedger; running && ledger > endLedger; ledger -= batchSize {
		currentFirst := ledger - batchSize + 1
		go fetchLedgers(currentFirst, ledger, filepath, context)
	inner:
		for {
			select {
			case <-context.done:
				break inner
			case <-interrupt:
				close(context.stop)
				running = false
				log.Println("interrupt")
			}
		}
	}
}
