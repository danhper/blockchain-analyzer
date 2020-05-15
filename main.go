package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danhper/blockchain-data-fetcher/core"
	"github.com/danhper/blockchain-data-fetcher/eos"
	"github.com/danhper/blockchain-data-fetcher/xrp"
	"github.com/urfave/cli/v2"
)

func addFetchFlags(flags []cli.Flag) []cli.Flag {
	return append(flags,
		&cli.StringFlag{
			Name:     "filepath, f",
			Value:    "",
			Usage:    "Base output filepath",
			Required: true,
		},
		&cli.IntFlag{
			Name:     "start, s",
			Required: true,
			Usage:    "Start block/ledger index",
		},
		&cli.IntFlag{
			Name:     "end, e",
			Required: true,
			Usage:    "End block/ledger index",
		},
	)
}

func blockchainFromCLI(c *cli.Context) (core.Blockchain, error) {
	if c.NArg() == 0 {
		return nil, fmt.Errorf("missing blockchain argument")
	}
	switch c.Args().Get(0) {
	case "xrp":
		return xrp.New(), nil
	case "eos":
		return eos.New(), nil
	default:
		return nil, fmt.Errorf("wrong blockchain argument. valid: 'xrp', 'eos'")
	}
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:      "fetch",
				Flags:     addFetchFlags(nil),
				Usage:     "Fetches blockchain data",
				ArgsUsage: "(eos | xrp)",
				Action: func(c *cli.Context) error {
					blockchain, err := blockchainFromCLI(c)
					if err != nil {
						cli.ShowSubcommandHelp(c)
						return cli.NewExitError(err.Error(), 1)
					}
					return blockchain.FetchData(c.String("filepath"), c.Uint64("start"), c.Uint64("end"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
