package main

import (
	"flag"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var start = flag.Int("start", 55406491, "start ledger index")
var end = flag.Int("end", 1, "last ledger index")
var filepath = flag.String("filepath", "", "base file path")

func addCommonFlags(flags []cli.Flag) []cli.Flag {
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
				Name:  "fetch-xrp",
				Flags: addCommonFlags([]cli.Flag{}),
				Usage: "Fetches XRP data",
				Action: func(c *cli.Context) error {
					return fetchXRPData(c.String("filepath"), c.Uint64("start"), c.Uint64("end"))
				},
			},
			{
				Name:  "fetch-eos",
				Flags: addCommonFlags([]cli.Flag{}),
				Usage: "Fetches EOS data",
				Action: func(c *cli.Context) error {
					return fetchEOSData(c.String("filepath"), c.Uint64("start"), c.Uint64("end"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
