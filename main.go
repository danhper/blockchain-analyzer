package main

import (
	"log"
	"os"

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

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:      "fetch",
				Flags:     addFetchFlags(nil),
				Usage:     "Fetches blockchain data",
				ArgsUsage: "(eos | xrp)",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						cli.ShowSubcommandHelp(c)
						return cli.NewExitError("missing blockchain argument", 1)
					}
					switch c.Args().Get(0) {
					case "xrp":
						return fetchXRPData(c.String("filepath"), c.Uint64("start"), c.Uint64("end"))
					case "eos":
						return fetchEOSData(c.String("filepath"), c.Uint64("start"), c.Uint64("end"))
					default:
						return cli.NewExitError("wrong blockchain argument. valid: 'xrp', 'eos'", 1)
					}
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
