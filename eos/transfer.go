package eos

import (
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/danhper/blockchain-analyzer/core"
	"github.com/danhper/blockchain-analyzer/processor"
)

type TransferData struct {
	From     string
	To       string
	Quantity string
	Memo     string
}

func parseTransferQuantity(rawQuantity string) (string, string, error) {
	tokens := strings.Split(rawQuantity, " ")
	if len(tokens) != 2 {
		return "", "", fmt.Errorf("expected 2 tokens, got %d", len(tokens))
	}
	if _, err := strconv.ParseFloat(tokens[0], 64); err != nil {
		return "", "", fmt.Errorf("quantity %s was not a valid float", tokens[0])
	}
	return tokens[0], tokens[1], nil
}

func ExportTransfers(globPattern string, start, end uint64, output string) error {
	writer, err := core.CreateFile(output)
	if err != nil {
		return err
	}
	defer writer.Close()
	csvWriter := csv.NewWriter(writer)

	headers := []string{
		"block",
		"tx",
		"account",
		"symbol",
		"from",
		"to",
		"quantity",
		"memo",
	}
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	blocks, err := processor.YieldAllBlocks(globPattern, New(), start, end)
	if err != nil {
		return err
	}
	for block := range blocks {
		eosBlock, ok := block.(*Block)
		if !ok {
			return err
		}

		for _, transaction := range eosBlock.Transactions {
			for _, action := range transaction.Trx.Transaction.Actions {
				if action.ActionName != "transfer" {
					continue
				}
				var transferData TransferData
				if err := fastJson.Unmarshal(action.Data, &transferData); err != nil {
					continue
				}
				quantity, symbol, err := parseTransferQuantity(transferData.Quantity)
				if err != nil {
					continue
				}

				row := []string{
					strconv.FormatUint(block.Number(), 10),
					transaction.Trx.Id,
					action.Account,
					symbol,
					transferData.From,
					transferData.To,
					quantity,
					transferData.Memo,
				}
				if err = csvWriter.Write(row); err != nil {
					log.Printf("could not write row: %s", err.Error())
				}
			}
		}
	}
	return nil
}
