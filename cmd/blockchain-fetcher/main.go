package main

import (
	"log"
	"os"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/danhper/blockchain-data-fetcher/eos"
	"github.com/danhper/blockchain-data-fetcher/processor"
	"github.com/danhper/blockchain-data-fetcher/xrp"
	"github.com/urfave/cli/v2"
)

func addFetchFlags(flags []cli.Flag) []cli.Flag {
	return append(flags,
		&cli.StringFlag{
			Name:     "filepath",
			Aliases:  []string{"f"},
			Value:    "",
			Usage:    "Base output filepath",
			Required: true,
		},
		&cli.IntFlag{
			Name:     "start",
			Aliases:  []string{"s"},
			Required: true,
			Usage:    "Start block/ledger index",
		},
		&cli.IntFlag{
			Name:     "end",
			Aliases:  []string{"e"},
			Required: true,
			Usage:    "End block/ledger index",
		},
	)
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
	default:
		cli.ShowSubcommandHelp(c)
		return nil, cli.NewExitError("wrong blockchain argument. valid: 'xrp', 'eos'", 1)
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
					return blockchain.FetchData(c.String("filepath"), c.Uint64("start"), c.Uint64("end"))
				},
			},
			{
				Name: "check",
				Flags: addBlockchainFlag([]cli.Flag{
					&cli.StringFlag{
						Name:     "pattern",
						Aliases:  []string{"p"},
						Value:    "",
						Usage:    "Patterns of files to check",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Value:    "",
						Usage:    "Blocks output path",
						Required: true,
					},
				}),
				Usage: "Checks for missing blocks in data",
				Action: func(c *cli.Context) error {
					blockchain, err := blockchainFromCLI(c)
					if err != nil {
						return err
					}
					return processor.OutputAllMissingBlockNumbers(blockchain, c.String("pattern"), c.String("output"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
