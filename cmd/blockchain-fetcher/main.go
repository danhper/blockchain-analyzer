package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/danhper/blockchain-data-fetcher/eos"
	"github.com/danhper/blockchain-data-fetcher/processor"
	"github.com/danhper/blockchain-data-fetcher/tezos"
	"github.com/danhper/blockchain-data-fetcher/xrp"
	"github.com/urfave/cli/v2"
)

func addStartFlag(flags []cli.Flag, required bool) []cli.Flag {
	return append(flags, &cli.IntFlag{
		Name:     "start",
		Aliases:  []string{"s"},
		Required: required,
		Value:    0,
		Usage:    "Start block/ledger index",
	})
}

func addEndFlag(flags []cli.Flag, required bool) []cli.Flag {
	return append(flags, &cli.IntFlag{
		Name:     "end",
		Aliases:  []string{"e"},
		Required: required,
		Value:    0,
		Usage:    "End block/ledger index",
	})
}

func addOutputFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:     "output",
		Aliases:  []string{"o"},
		Usage:    "Base output filepath",
		Required: true,
	})
}

func addFetchFlags(flags []cli.Flag) []cli.Flag {
	return addStartFlag(addEndFlag(addOutputFlag(flags), true), true)
}

func addPatternFlag(flags []cli.Flag) []cli.Flag {
	return append(flags, &cli.StringFlag{
		Name:     "pattern",
		Aliases:  []string{"p"},
		Value:    "",
		Usage:    "Patterns of files to check",
		Required: true,
	})
}

func addBlockchainFlag(flags []cli.Flag) []cli.Flag {
	return append(flags,
		&cli.StringFlag{
			Name:     "blockchain",
			Aliases:  []string{"b"},
			Value:    "",
			Usage:    "Blockchain to use",
			Required: true,
		},
	)
}

func blockchainFromCLI(c *cli.Context) (core.Blockchain, error) {
	switch c.String("blockchain") {
	case "xrp":
		return xrp.New(), nil
	case "eos":
		return eos.New(), nil
	case "tezos":
		return tezos.New(), nil
	default:
		cli.ShowSubcommandHelp(c)
		return nil, cli.NewExitError("wrong blockchain argument. valid: 'xrp', 'eos', 'tezos'", 1)
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "fetch",
				Flags: addFetchFlags(addBlockchainFlag(nil)),
				Usage: "Fetches blockchain data",
				Action: func(c *cli.Context) error {
					blockchain, err := blockchainFromCLI(c)
					if err != nil {
						return err
					}
					return blockchain.FetchData(c.String("output"), c.Uint64("start"), c.Uint64("end"))
				},
			},
			{
				Name:  "check",
				Flags: addBlockchainFlag(addStartFlag(addOutputFlag(addPatternFlag(nil)), true)),
				Usage: "Checks for missing blocks in data",
				Action: func(c *cli.Context) error {
					blockchain, err := blockchainFromCLI(c)
					if err != nil {
						return err
					}
					return processor.OutputAllMissingBlockNumbers(
						blockchain, c.String("pattern"), c.String("output"), c.Uint64("start"))
				},
			},
			{
				Name: "count-transactions",
				Flags: addBlockchainFlag(addEndFlag(
					addStartFlag(addPatternFlag(nil), false), false)),
				Usage: "Count the number of transactions in the data",
				Action: func(c *cli.Context) error {
					blockchain, err := blockchainFromCLI(c)
					if err != nil {
						return err
					}
					count, err := processor.CountTransactions(
						blockchain, c.String("pattern"),
						c.Uint64("start"), c.Uint64("end"))
					if err != nil {
						return err
					}
					fmt.Printf("found %d transactions\n", count)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
